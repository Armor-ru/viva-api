package viva_api

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/tplext"
	httpTransport "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/Armor-ru/viva-api/internal/vivaclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"

	"github.com/spf13/cast"
)

type Viva struct {
	intTransport types.Transport
	extTransport types.Transport

	secrets []string
	smpp    SmppConfig

	testTariffs []string
	channels    Channels

	SmsTpl     *template.Template
	smppSender *SmppSender

	accountId  string
	vivaClient *vivaclient.Client
}

func (s *Viva) InitHandlers() {
	if s.extTransport != nil {
		s.initMiddleWare()

		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", s.ExtAppPartnerProductActivationRequestHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", s.ExtAppPartnerProductActivationHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", s.ExtAppPartnerProductRemoveHandler)
		s.extTransport.Subscribe("POST /landing/init-subscription", s.LandingInitHandler)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.LandingConfirmHandler)
	}

	if s.intTransport != nil {
		s.intTransport.Subscribe("order/completed", s.onCompletedHandler)
	}

	tplText := s.smpp.Template
	if tplText == "" {
		tplText = "{{.ProductName}}{{ if .Quantity }} for {{ pluralizeEn .Quantity \"device\" \"devices\" }}{{ end }}\n" +
			"Activation code: {{.ActivationCode}}\n" +
			"Download link: {{.DownloadURL}}"
	}
	SmsTpl, _ := template.New("sms").Funcs(tplext.Funcs).Parse(tplText)
	s.SmsTpl = SmsTpl

	s.smppSender = NewSmppSender(s.smpp)
}

func (s *Viva) initMiddleWare() {
	signature := httpTransport.Signature(httpTransport.SignatureConfig{
		Secrets: s.secrets,
		Header:  "X-Signature",
	})

	s.extTransport.Middleware(cors.New(cors.Config{
		Next: func(c *fiber.Ctx) bool {
			return !isLandingPath(c.Path())
		},
		AllowOrigins: "*",
		AllowMethods: "POST,OPTIONS",
		AllowHeaders: "Content-Type,Accept,X-MSISDN,X-Msisdn,X-Phone-Number",
	}))
	s.extTransport.Middleware(func(c *fiber.Ctx) error {
		if isLandingPath(c.Path()) {
			return c.Next()
		}
		return signature(c)
	})
}

func isLandingPath(path string) bool {
	return path == "/landing" || strings.HasPrefix(path, "/landing/")
}

func (s *Viva) ExtAppPartnerProductActivationRequestHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeNew)
}

func (s *Viva) ExtAppPartnerProductActivationHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeRenew)
}

func (s *Viva) ExtAppPartnerProductRemoveHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeCancel)
}

func (s *Viva) handleCreate(ctx types.HandlerContext, orderType types.OrderType) {
	data := ExtReq{}
	ctx.Data(&data)

	if _, err := s.createOrder(orderType, data.PhoneNum, data.ProductCode); err != nil {
		logger.Error().Msg("can not create order, " + err.Error())
		return
	}

	ctx.Response("")
}

func (s *Viva) createOrder(orderType types.OrderType, phone, externalID string) (string, error) {
	if s.intTransport == nil {
		return "", fmt.Errorf("intTransport is not configured")
	}

	phone = strings.TrimSpace(phone)
	externalID = strings.TrimSpace(externalID)
	if phone == "" || externalID == "" {
		return "", fmt.Errorf("phoneNum and productCode are required")
	}

	orderId := uuid.NewSHA1(uuid.MustParse(s.accountId), []byte(externalID+":"+phone)).String()

	// Наполнение items для новых и orderId для старых заказов
	var items []types.OrderItemRequest
	if orderType == types.OrderTypeNew {
		items = append(items, types.OrderItemRequest{
			Id:         uuid.NewString(),
			ExternalId: &externalID,
		})
	}

	// Формируем тело заказа
	newOrder := types.OrderCreateRequest{
		Id:   orderId,
		Type: orderType,
		Fields: types.JSON{
			"phone": phone,
		},
		Items: items,
	}

	// Создаем заказ
	_, err := s.intTransport.Send("order/create", newOrder, types.SendOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("send order/create: %w", err)
	}

	return orderId, nil
}

func (s *Viva) onCompletedHandler(ctx types.HandlerContext) {
	order := types.OrderResponse{}
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.completed")

	if order.Status == "error" {
		logger.Info().Str("orderId", order.ID).Msg("order status error, no need to send notify")
		return
	}

	if len(order.Items) == 0 {
		logger.Error().Str("orderId", order.ID).Msg("can not send notify, order items is empty")
		return
	}

	hasSendSms := false
	for _, it := range order.Items {
		if it.Type == "activate" || it.Type == "reactivate" {
			hasSendSms = true
			break
		}
	}

	if !hasSendSms {
		logger.Info().Str("orderId", order.ID).Msg("no need to send notify with activation code")
		return
	}

	item := order.Items[0]

	activationCode := strings.TrimSpace(cast.ToString(item.Artifacts["ActivationCode"]))
	if activationCode == "" {
		logger.Error().Str("orderId", order.ID).Msg("can not send notify, ActivationCode is empty")
		return
	}

	// Raw Download
	rawDownload, ok := item.Artifacts["download"].([]interface{})
	if !ok {
		rawDownload = []interface{}{}
	}

	// download convert
	downloads := make([]map[string]interface{}, 0, len(rawDownload))
	for _, v := range rawDownload {
		if m, ok := v.(map[string]interface{}); ok {
			downloads = append(downloads, m)
		}
	}

	if len(downloads) == 0 {
		logger.Error().Str("orderId", order.ID).Msg("can not send notify, artifacts download not found")
		return
	}

	downloadURL := strings.TrimSpace(cast.ToString(downloads[0]["url"]))
	if downloadURL == "" {
		logger.Error().Str("orderId", order.ID).Msg("can not send notify, DownloadURL is empty")
		return
	}

	// productName := cast.ToString(item.Product.Name)
	// if productName == "" {
	// 	logger.Error().Str("orderId", order.ID).Msg("can not send notify, product name is empty")
	// }

	smsData := SmsData{
		ProductName:    cast.ToString(item.Product.Name),
		Quantity:       0,
		ActivationCode: activationCode,
		DownloadURL:    downloadURL,
	}

	if quantity, ok := item.Options["quantity"]; ok {
		smsData.Quantity = cast.ToInt(quantity)
	}

	var smsMessage bytes.Buffer
	if err := s.SmsTpl.Execute(&smsMessage, smsData); err != nil {
		logger.Error().Str("orderId", order.ID).Msg("render notify template failed, " + err.Error())
		return
	}

	rawPhone, ok := order.Fields["phone"]
	if !ok {
		logger.Error().Str("orderId", order.ID).Interface("fields", order.Fields).Msg("can not send notify, order has no phone")
		return
	}

	phone := strings.TrimSpace(cast.ToString(rawPhone))
	if phone == "" {
		logger.Error().Str("orderId", order.ID).Msg("can not send notify, phone is empty")
		return
	}

	logger.Info().Str("orderId", order.ID).Str("phone", phone).Str("text", smsMessage.String()).Msg("send notify with activation code and download link")

	// Разделяем данные для grafana
	if err := s.smppSender.Send(phone, smsMessage.String()); err != nil {
		smppErr := err.(*SmppError)

		log := logger.Error().Str("orderId", order.ID)

		for k, v := range smppErr.Fields {
			log = log.Interface(k, v)
		}

		log.Msg("send smpp notify failed, " + err.Error())
		return
	}

	logger.Info().Str("orderId", order.ID).Str("phone", phone).Msg("notify sent successfully")
}
