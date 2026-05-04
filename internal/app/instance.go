package viva_api

import (
	"strings"
	"text/template"
	"time"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/tplext"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/Armor-ru/sds-go/pkg/types"
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

	vivaPartner        PartnerSubscriptionAPI
	defaultProductName string
	// orderProductCode — externalId для order/create (тот же UUID, что productCode во вебхуке ActivationRequest).
	orderProductCode string

	accountId string
}

func (s *Viva) InitHandlers() {
	if s.extTransport != nil {
		s.initMiddleWare()

		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", s.ExtAppPartnerProductActivationRequestHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", s.ExtAppPartnerProductActivationHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", s.ExtAppPartnerProductRemoveHandler)

		s.extTransport.Subscribe("POST /landing/init-subscription", s.LandingInitSubscriptionHandler)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.LandingConfirmSubscriptionHandler)
		s.extTransport.Subscribe("GET /landing/subscriber-info/:phoneNum", s.LandingGetSubscriberInfoGETHandler)
		s.extTransport.Subscribe("POST /landing/subscriber-info", s.LandingGetSubscriberInfoPOSTHandler)

		s.extTransport.Subscribe("POST /landing/:locale/init-subscription", s.LandingInitSubscriptionLocalizedHandler)
		s.extTransport.Subscribe("POST /landing/:locale/confirm-subscription", s.LandingConfirmSubscriptionLocalizedHandler)
		s.extTransport.Subscribe("GET /landing/:locale/subscriber-info/:phoneNum", s.LandingGetSubscriberInfoGETLocalizedHandler)
		s.extTransport.Subscribe("POST /landing/:locale/subscriber-info", s.LandingGetSubscriberInfoPOSTLocalizedHandler)
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
	ht, ok := s.extTransport.(*httpt.Transport)
	sig := httpt.Signature(httpt.SignatureConfig{Secrets: s.secrets, Header: "X-Signature"})
	if !ok {
		logger.Warn().Msg("extTransport is not *http.Transport; webhook signature middleware skipped")
		return
	}
	// Лендинг со статики (другой порт / CDN) — браузер шлёт cross-origin POST.
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
	s.sendOrderCreate(orderType, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
	_ = ctx.Response("")
}

// sendOrderCreate шлёт order/create в NATS (как обработчик вебхука ExtAppPartneerProductActivationRequest и др.).
func (s *Viva) sendOrderCreate(orderType types.OrderType, phone, externalID, smsScenario, smsLocale string) {
	if s.intTransport == nil {
		logger.Warn().Msg("intTransport nil, skip order/create")
		return
	}
	phone = strings.TrimSpace(phone)
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		logger.Warn().Str("phone", phone).Str("orderType", string(orderType)).Msg("skip order/create: productCode empty")
		return
	}
	if phone == "" {
		logger.Warn().Msg("skip order/create: phone empty")
		return
	}

	orderId := uuid.NewSHA1(uuid.MustParse(s.accountId), []byte(externalID+":"+phone)).String()

	var items []types.OrderItemRequest
	if orderType == types.OrderTypeNew {
		items = append(items, types.OrderItemRequest{
			Id:         uuid.NewString(),
			ExternalId: &externalID,
		})
	}

	wh := ""
	switch orderType {
	case types.OrderTypeNew:
		wh = WHActivationReq
	case types.OrderTypeRenew:
		wh = WHActivation
	case types.OrderTypeCancel:
		wh = WHRemove
	}

	newOrder := types.OrderCreateRequest{
		Id:   orderId,
		Type: orderType,
		Fields: types.JSON{
			"phone": phone,
		},
		CustomData: types.JSON{
			CDVivaWebhook:   wh,
			CDSmsScenario:   strings.TrimSpace(smsScenario),
			CDSmsLocale:     LocaleOrDefault(smsLocale),
		},
		Items: items,
	}

	_, err := s.intTransport.Send("order/create", newOrder, types.SendOptions{
		Timeout: 3 * time.Second,
	})
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
		logger.Error().Str("orderId", order.ID).Interface("fields", order.Fields).Msg("can not send notify, order has no phone")
		return
	}

	wh := strings.TrimSpace(cast.ToString(order.CustomData[CDVivaWebhook]))
	scenarioHint := strings.TrimSpace(cast.ToString(order.CustomData[CDSmsScenario]))
	locale := orderSmsLocale(order)

	switch wh {
	case WHActivation:
		s.sendRenewSms(order, phone, scenarioHint, locale)
	case WHRemove:
		s.sendRemoveSms(order, phone, scenarioHint, locale)
	default:
		s.sendNewActivationSms(order, phone, locale)
	}
}

func orderSmsLocale(order types.OrderResponse) string {
	return LocaleOrDefault(cast.ToString(order.CustomData[CDSmsLocale]))
}

func orderPhone(order types.OrderResponse) string {
	raw, ok := order.Fields["phone"]
	if !ok {
		return ""
	}
	return strings.TrimSpace(cast.ToString(raw))
}

func (s *Viva) sendNewActivationSms(order types.OrderResponse, phone, locale string) {
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

	rawDownload, ok := item.Artifacts["download"].([]interface{})
	if !ok {
		rawDownload = []interface{}{}
	}
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

	payload := SmsScenarioPayload{
		ProductLabel: cast.ToString(item.Product.Name),
		LicenseKey:   activationCode,
		DownloadURL:  downloadURL,
		Locale:       locale,
	}
	if err := s.smppSendScenario(phone, Sms2Welcome, payload); err != nil {
		logger.Error().Str("orderId", order.ID).Err(err).Msg("SMPP sms2")
		return
	}
	if err := s.smppSendScenario(phone, Sms3License, payload); err != nil {
		logger.Error().Str("orderId", order.ID).Err(err).Msg("SMPP sms3")
		return
	}
	logger.Info().Str("orderId", order.ID).Str("phone", phone).Msg("SPEC §8 sms2+sms3 sent")
}

func (s *Viva) sendRenewSms(order types.OrderResponse, phone, scenarioHint, locale string) {
	sc := Sms4Paid
	switch SmsScenario(scenarioHint) {
	case Sms5TrialRemind, Sms15Booster, Sms4Paid:
		sc = SmsScenario(scenarioHint)
	}
	payload := SmsScenarioPayload{ProductLabel: productLabelFromOrder(order), Locale: locale}
	if sc == Sms5TrialRemind && order.EndTime != nil {
		payload.TrialEndDate = order.EndTime.Format("02.01.2006")
	}
	if err := s.smppSendScenario(phone, sc, payload); err != nil {
		logger.Error().Str("orderId", order.ID).Err(err).Str("scenario", string(sc)).Msg("SMPP renew webhook")
	}
}

func (s *Viva) sendRemoveSms(order types.OrderResponse, phone, scenarioHint, locale string) {
	sc := SmsServiceRemoved
	if SmsScenario(scenarioHint) == Sms14NoFunds {
		sc = Sms14NoFunds
	}
	payload := SmsScenarioPayload{ProductLabel: productLabelFromOrder(order), Locale: locale}
	if err := s.smppSendScenario(phone, sc, payload); err != nil {
		logger.Error().Str("orderId", order.ID).Err(err).Str("scenario", string(sc)).Msg("SMPP remove webhook")
	}
}

func productLabelFromOrder(order types.OrderResponse) string {
	if len(order.Items) == 0 {
		return ""
	}
	return cast.ToString(order.Items[0].Product.Name)
}
