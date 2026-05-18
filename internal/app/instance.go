package viva_api

import (
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
		// ExtAppPartneerProductActivationRequest stay for test because have method "init" which make same things
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
	s.handleExtWebhook(ctx, types.OrderTypeNew)
}

func (s *Viva) ExtAppPartnerProductActivationHandler(ctx types.HandlerContext) {
	s.handleExtWebhook(ctx, types.OrderTypeRenew)
}

func (s *Viva) ExtAppPartnerProductRemoveHandler(ctx types.HandlerContext) {
	s.handleExtWebhook(ctx, types.OrderTypeCancel)
}

func (s *Viva) handleExtWebhook(ctx types.HandlerContext, orderType types.OrderType) {
	data := ExtReq{}
	ctx.Data(&data)
	s.sendOrderCreate(orderType, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
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

	phone := orderPhone(order)
	if phone == "" {
		logger.Error().Str("orderId", order.ID).Msg("can not send notify, order has no phone")
		return
	}

	wh := strings.TrimSpace(cast.ToString(order.CustomData[cdVivaWebhook]))
	scenario := strings.TrimSpace(cast.ToString(order.CustomData[cdSmsScenario]))
	locale := localeOrDefault(cast.ToString(order.CustomData[cdSmsLocale]))

	switch wh {
	case whActivation:
		sc := "sms4"
		if scenario == "sms5" || scenario == "sms15" || scenario == "sms4" {
			sc = scenario
		}
		s.sendScenarioSMS(order, phone, locale, sc)
	case whRemove:
		sc := "sms_deactivated"
		if scenario == "sms14" {
			sc = "sms14"
		}
		s.sendScenarioSMS(order, phone, locale, sc)
	default:
		s.sendActivationSMS(order, phone, locale)
	}
}

func orderPhone(order types.OrderResponse) string {
	raw, ok := order.Fields["phone"]
	if !ok {
		return ""
	}
	return strings.TrimSpace(cast.ToString(raw))
}

func (s *Viva) sendActivationSMS(order types.OrderResponse, phone, locale string) {
	if len(order.Items) == 0 {
		return
	}
	ok := false
	for _, it := range order.Items {
		if it.Type == "activate" || it.Type == "reactivate" {
			ok = true
			break
		}
	}
	if !ok {
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

	label := cast.ToString(item.Product.Name)
	d := SmsData{ProductLabel: label, ActivationCode: code, LicenseKey: code, DownloadURL: url}
	if !s.sendScenarioSMS(order, phone, locale, "sms2", d) {
		return
	}
	s.sendScenarioSMS(order, phone, locale, "sms3", d)
}

func (s *Viva) sendScenarioSMS(order types.OrderResponse, phone, locale, scenario string, data ...SmsData) bool {
	d := SmsData{ProductLabel: productLabel(order)}
	if order.EndTime != nil {
		d.TrialEndDate = order.EndTime.Format("02.01.2006")
	}
	if len(data) > 0 {
		d = data[0]
		if d.ProductLabel == "" {
			d.ProductLabel = productLabel(order)
		}
	}

	body := smsText(scenario, locale, d)
	if body == "" {
		return true
	}
	if err := s.smppSender.Send(phone, body); err != nil {
		logSmppErr(order.ID, err)
		return false
	}
	logger.Info().Str("orderId", order.ID).Str("scenario", scenario).Str("phone", phone).Msg("SMS sent")
	return true
}

func productLabel(order types.OrderResponse) string {
	if len(order.Items) == 0 {
		return ""
	}
	return cast.ToString(order.Items[0].Product.Name)
}

func logSmppErr(orderID string, err error) {
	smppErr, ok := err.(*SmppError)
	log := logger.Error().Str("orderId", orderID)
	if ok {
		for k, v := range smppErr.Fields {
			log = log.Interface(k, v)
		}
	}
	log.Msg("send smpp notify failed, " + err.Error())
}
