package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
	"github.com/spf13/cast"
)

// Сценарий 4: триал закончился → платная подписка (СМС №4).
// Шаги 1–6 (viva-api): webhook Viva → order/create type=renew.
// Шаги 7–8 (SDS): обработка order/create → публикация order/completed в NATS.
// Шаги 9–11 (viva-api): приём order/completed → каталог → СМС №4 по SMPP.

// Viva вызывает webhook при старте платного периода (обычно сразу после триала).
const extAppPartnerProductActivationPath = "ExtAppPartneerProductActivation"

const orderCompletedSubscribePath = "order/completed"

// handleExtAppPartnerProductActivation — вход сценария 4: конец триала → продление подписки.
func (s *viva) handleExtAppPartnerProductActivation(ctx types.HandlerContext) {
	data := ExtReq{}
	ctx.Data(&data)

	if err := scenario4Step1(data); err != nil {
		logger.Error().Err(err).
			Str("webhook", extAppPartnerProductActivationPath).
			Msg("scenario 4 step 1: invalid Viva webhook payload")
		_ = ctx.Response("")
		return
	}

	if err := s.scenario4Step2ProcessWebhook(data); err != nil {
		logger.Error().Err(err).
			Str("webhook", extAppPartnerProductActivationPath).
			Str("phone", normalizePhone(data.PhoneNum)).
			Str("productCode", strings.TrimSpace(data.ProductCode)).
			Msg("scenario 4 step 2: webhook processing failed")
		_ = ctx.Response("")
		return
	}

	_ = ctx.Response("")
}

// scenario4Step1: Viva → POST ExtAppPartneerProductActivation (переход триал → платный период).
func scenario4Step1(data ExtReq) error {
	phone := normalizePhone(data.PhoneNum)
	productCode := strings.TrimSpace(data.ProductCode)
	if phone == "" {
		return fmt.Errorf("phoneNum is required")
	}
	if productCode == "" {
		return fmt.Errorf("productCode is required")
	}

	logger.Info().
		Str("webhook", extAppPartnerProductActivationPath).
		Str("phone", phone).
		Str("productCode", productCode).
		Msg("scenario 4 step 1: received ExtAppPartneerProductActivation from Viva (trial ended / subscription renew)")

	return nil
}

// scenario4Step2ProcessWebhook продолжает сценарий 4 после приёма webhook.
func (s *viva) scenario4Step2ProcessWebhook(data ExtReq) error {
	logger.Info().
		Str("phone", normalizePhone(data.PhoneNum)).
		Str("productCode", strings.TrimSpace(data.ProductCode)).
		Msg("scenario 4 step 2: processing ExtAppPartneerProductActivation")

	product, err := s.productForExternalID(FlowPaid, "3", strings.TrimSpace(data.ProductCode))
	if err != nil {
		return err
	}

	orderID, err := s.scenario4Step4FormOrderID(data.PhoneNum, product)
	if err != nil {
		return err
	}

	req := s.scenario4Step5ConvertToRenewOrderCreate(orderID, data.PhoneNum, product.GetExternalID())
	return s.scenario4Step6PublishOrderCreate(req)
}

// scenario4Step6PublishOrderCreate публикует order/create в NATS (обрабатывает SDS).
func (s *viva) scenario4Step6PublishOrderCreate(req types.OrderCreateRequest) error {
	if s.intTransport == nil {
		return fmt.Errorf("intTransport is not configured")
	}

	logger.Info().
		Str("orderId", req.ID).
		Str("type", string(req.Type)).
		Str("topic", "order/create").
		Msg("scenario 4 step 6: publishing order/create to NATS")

	if err := s.publishOrderCreate(req); err != nil {
		logger.Error().Err(err).
			Str("orderId", req.ID).
			Msg("scenario 4 step 6: order/create publish failed")
		return err
	}

	logger.Info().
		Str("orderId", req.ID).
		Str("type", string(req.Type)).
		Str("topic", "order/create").
		Msg("scenario 4 step 6: order/create published to NATS")

	return nil
}

// scenario4Step5ConvertToRenewOrderCreate преобразует поля webhook в order/create (type=renew).
func (s *viva) scenario4Step5ConvertToRenewOrderCreate(orderID, phone, productCode string) types.OrderCreateRequest {
	phone = normalizePhone(phone)
	productCode = strings.TrimSpace(productCode)
	lang := s.resolveLang(phone, "")

	req := buildOrderCreateRequest(types.OrderTypeRenew, orderID, phone, productCode, lang)
	logger.Info().
		Str("orderId", orderID).
		Str("phone", phone).
		Str("productCode", productCode).
		Str("lang", lang).
		Str("type", string(types.OrderTypeRenew)).
		Msg("scenario 4 step 5: converted webhook data to order/create (type=renew)")

	return req
}

// scenario4Step4FormOrderID формирует стабильный orderId из externalId продукта и телефона абонента.
func (s *viva) scenario4Step4FormOrderID(phone string, product catalog.Product) (string, error) {
	phone = normalizePhone(phone)
	if phone == "" {
		return "", fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(s.accountId) == "" {
		return "", fmt.Errorf("accountId is not configured")
	}

	productCode := strings.TrimSpace(product.GetExternalID())
	if productCode == "" {
		return "", fmt.Errorf("product externalId is empty")
	}

	orderID := s.getOrderId(phone, productCode)
	logger.Info().
		Str("orderId", orderID).
		Str("phone", phone).
		Str("productCode", productCode).
		Msg("scenario 4 step 4: orderId formed from product and phone")

	return orderID, nil
}

func isScenario4OrderCompleted(order types.OrderResponse) bool {
	if strings.TrimSpace(order.ID) == "" {
		return false
	}
	for _, item := range order.Items {
		if item.Type == "renew" {
			return true
		}
	}
	return false
}

// handleScenario4OrderCompleted выполняет сценарий 4 после order/completed от SDS (шаги 9+).
func (s *viva) handleScenario4OrderCompleted(order types.OrderResponse) error {
	if err := scenario4Step9ReceiveOrderCompleted(order); err != nil {
		return err
	}

	externalID := productCodeFromOrderResponse(order)
	product, err := s.productForExternalID(FlowPaid, "10", externalID)
	if err != nil {
		return err
	}

	return s.scenario4Step11SendSMS4(order, product)
}

// scenario4Step11SendSMS4 отправляет СМС №4 (триал закончился → платная версия) по SMPP.
func (s *viva) scenario4Step11SendSMS4(order types.OrderResponse, product catalog.Product) error {
	phone := phoneFromOrderResponse(order)
	if phone == "" {
		return fmt.Errorf("order/completed: phone is empty")
	}

	productCode := strings.TrimSpace(product.GetExternalID())
	if s.paidWelcomeStore().AlreadySent(phone, productCode) {
		logger.Info().
			Str("phone", phone).
			Str("productCode", productCode).
			Msg("scenario 4 step 12: SMS #4 already sent after trial→paid, skip on repeat renew")
		return nil
	}

	lang := ""
	if order.CustomData != nil {
		lang = cast.ToString(order.CustomData["lang"])
	}
	lang = s.resolveLang(phone, lang)
	if lang == "" {
		lang = normalizeLang(product.GetDefaultLanguage())
	}

	data := map[string]interface{}{
		"Phone":      phone,
		"ExternalID": productCode,
		"Language":   lang,
	}

	text := product.GetNotify(NotifyWelcomePaid, data, lang)
	if text == "" {
		return fmt.Errorf("welcome_paid notification template is empty")
	}

	logger.Info().
		Str("phone", phone).
		Str("lang", lang).
		Msg("scenario 4 step 11: sending SMS #4 (trial over, paid activated) via SMPP")

	if err := s.notify(phone, text); err != nil {
		return fmt.Errorf("send welcome_paid sms: %w", err)
	}

	s.paidWelcomeStore().MarkSent(phone, productCode)
	return nil
}

// scenario4Step9ReceiveOrderCompleted: подписчик viva-api получает order/completed из NATS.
func scenario4Step9ReceiveOrderCompleted(order types.OrderResponse) error {
	if strings.TrimSpace(order.ID) == "" {
		return fmt.Errorf("order/completed: order id is empty")
	}
	if strings.TrimSpace(order.Status) == "" {
		return fmt.Errorf("order/completed: status is empty")
	}

	logger.Info().
		Str("orderId", order.ID).
		Str("status", order.Status).
		Str("topic", orderCompletedSubscribePath).
		Str("natsSubject", "order.completed").
		Msg("scenario 4 step 9: viva-service received order/completed from NATS")

	return nil
}
