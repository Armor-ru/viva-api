package viva_api

import (
	"fmt"
	"strings"
	"time"

	"github.com/Armor-ru/sds-go/pkg/logger"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/spf13/cast"

	"github.com/Armor-ru/viva-api/internal/vivaclient"
)

type Viva struct {
	intTransport types.Transport
	extTransport types.Transport

	secrets    []string
	smpp       SmppConfig
	accountId  string
	vivaClient *vivaclient.Client

	smppSender *SmppSender
}

func (s *Viva) InitHandlers() {
	s.smppSender = NewSmppSender(s.smpp)

	if s.extTransport != nil {
		s.initMiddleWare()
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", s.ExtAppPartnerProductActivationRequestHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", s.ExtAppPartnerProductActivationHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", s.ExtAppPartnerProductRemoveHandler)

		s.extTransport.Subscribe("POST /landing/init-subscription", s.LandingInitHandler)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.LandingConfirmHandler)
		s.extTransport.Subscribe("POST /landing/:locale/init-subscription", s.LandingInitHandler)
		s.extTransport.Subscribe("POST /landing/:locale/confirm-subscription", s.LandingConfirmHandler)
	}

	if s.intTransport != nil {
		s.intTransport.Subscribe("order/completed", s.onCompletedHandler)
	}
}

func (s *Viva) initMiddleWare() {
	ht, ok := s.extTransport.(*httpt.Transport)
	sig := httpt.Signature(httpt.SignatureConfig{Secrets: s.secrets, Header: "X-Signature"})
	if !ok {
		logger.Warn().Msg("extTransport is not *http.Transport; signature middleware skipped")
		return
	}
	ht.Middleware(cors.New(cors.Config{
		Next: func(c *fiber.Ctx) bool {
			return !strings.HasPrefix(c.Path(), "/landing/")
		},
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Accept",
	}))
	ht.Middleware(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/landing/") {
			return c.Next()
		}
		return sig(c)
	})
}

func (s *Viva) ExtAppPartnerProductActivationRequestHandler(ctx types.HandlerContext) {
	data := ExtReq{}
	ctx.Data(&data)
	s.sendOrderCreate(types.OrderTypeNew, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
	_ = ctx.Response("")
}

func (s *Viva) ExtAppPartnerProductActivationHandler(ctx types.HandlerContext) {
	data := ExtReq{}
	ctx.Data(&data)
	s.sendOrderCreate(types.OrderTypeRenew, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
	_ = ctx.Response("")
}

func (s *Viva) ExtAppPartnerProductRemoveHandler(ctx types.HandlerContext) {
	data := ExtReq{}
	ctx.Data(&data)
	s.sendOrderCreate(types.OrderTypeCancel, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
	_ = ctx.Response("")
}

func (s *Viva) handleCreate(ctx types.HandlerContext, orderType types.OrderType, phone, productCode, smsScenario, locale string) {
	s.sendOrderCreate(orderType, phone, productCode, smsScenario, locale)
}

func (s *Viva) sendOrderCreate(orderType types.OrderType, phone, externalID, smsScenario, locale string) {
	if s.intTransport == nil {
		logger.Warn().Msg("intTransport nil, skip order/create")
		return
	}
	phone = strings.TrimSpace(phone)
	externalID = strings.TrimSpace(externalID)
	if phone == "" || externalID == "" {
		logger.Warn().Msg("skip order/create: empty phone or productCode")
		return
	}

	orderId := uuid.NewSHA1(uuid.MustParse(s.accountId), []byte(externalID+":"+phone)).String()

	var items []types.OrderItemRequest
	if orderType == types.OrderTypeNew {
		ext := externalID
		items = append(items, types.OrderItemRequest{
			Id:         uuid.NewString(),
			ExternalId: &ext,
		})
	}

	wh := whActivationReq
	switch orderType {
	case types.OrderTypeRenew:
		wh = whActivation
	case types.OrderTypeCancel:
		wh = whRemove
	}

	newOrder := types.OrderCreateRequest{
		Id:   orderId,
		Type: orderType,
		Fields: types.JSON{
			"phone": phone,
		},
		CustomData: types.JSON{
			cdVivaWebhook: wh,
			cdSmsScenario: strings.TrimSpace(smsScenario),
			cdSmsLocale:   localeOrDefault(locale),
		},
		Items: items,
	}

	_, err := s.intTransport.Send("order/create", newOrder, types.SendOptions{Timeout: 3 * time.Second})
	if err != nil {
		logger.Error().Interface("payload", newOrder).Msg("can not create order, " + err.Error())
		return
	}
	logger.Info().Str("orderId", orderId).Str("type", string(orderType)).Str("phone", phone).Msg("order/create sent")
}

func (s *Viva) onCompletedHandler(ctx types.HandlerContext) {
	order := types.OrderResponse{}
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.completed")

	if order.Status == "error" {
		logger.Info().Str("orderId", order.ID).Msg("order status error, no need to send notify")
		return
	}

	rawPhone, ok := order.Fields["phone"]
	if !ok {
		logger.Error().Str("orderId", order.ID).Msg("can not send notify, order has no phone")
		return
	}
	phone := strings.TrimSpace(cast.ToString(rawPhone))
	if phone == "" {
		return
	}

	wh := strings.TrimSpace(cast.ToString(order.CustomData[cdVivaWebhook]))
	scenario := strings.TrimSpace(cast.ToString(order.CustomData[cdSmsScenario]))
	locale := localeOrDefault(cast.ToString(order.CustomData[cdSmsLocale]))

	label := ""
	if len(order.Items) > 0 {
		label = cast.ToString(order.Items[0].Product.Name)
	}
	trialEnd := ""
	if order.EndTime != nil {
		trialEnd = order.EndTime.Format("02.01.2006")
	}

	send := func(scenario string, args ...interface{}) bool {
		body := fmt.Sprintf(GetTemplate(scenario, locale), args...)
		if body == "" {
			return true
		}
		if err := s.smppSender.Send(phone, body); err != nil {
			smppErr, ok := err.(*SmppError)
			log := logger.Error().Str("orderId", order.ID)
			if ok {
				for k, v := range smppErr.Fields {
					log = log.Interface(k, v)
				}
			}
			log.Msg("send smpp notify failed, " + err.Error())
			return false
		}
		logger.Info().Str("orderId", order.ID).Str("scenario", scenario).Str("phone", phone).Msg("SMS sent")
		return true
	}

	switch wh {
	case whActivation:
		switch scenario {
		case "sms5":
			if trialEnd != "" {
				send("sms5_with_date", label, trialEnd)
			} else {
				send("sms5_soon", label)
			}
		case "sms15":
			send("sms15")
		default:
			send("sms4", label)
		}
	case whRemove:
		if scenario == "sms14" {
			send("sms14")
		} else {
			send("sms_deactivated", label)
		}
	default:
		if len(order.Items) == 0 {
			return
		}
		hasActivation := false
		for _, it := range order.Items {
			if it.Type == "activate" || it.Type == "reactivate" {
				hasActivation = true
				break
			}
		}
		if !hasActivation {
			logger.Info().Str("orderId", order.ID).Msg("no need to send notify with activation code")
			return
		}

		item := order.Items[0]
		code := strings.TrimSpace(cast.ToString(item.Artifacts["ActivationCode"]))
		if code == "" {
			logger.Error().Str("orderId", order.ID).Msg("can not send notify, ActivationCode is empty")
			return
		}

		rawDownload, _ := item.Artifacts["download"].([]interface{})
		downloads := make([]map[string]interface{}, 0)
		for _, v := range rawDownload {
			if m, ok := v.(map[string]interface{}); ok {
				downloads = append(downloads, m)
			}
		}
		if len(downloads) == 0 {
			logger.Error().Str("orderId", order.ID).Msg("can not send notify, artifacts download not found")
			return
		}
		url := strings.TrimSpace(cast.ToString(downloads[0]["url"]))
		if url == "" {
			logger.Error().Str("orderId", order.ID).Msg("can not send notify, DownloadURL is empty")
			return
		}

		product := cast.ToString(item.Product.Name)
		if !send("sms2", product) {
			return
		}
		send("sms3", product, code, url)
	}
}
