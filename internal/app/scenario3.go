package viva_api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
	"github.com/spf13/cast"
)

// Путь подписки NATS (sds-go ParsePath: order/expires → subject order.expires).
const orderExpiresSubscribePath = "order/expires"

// orderExpiresPayload — нормализованное событие order.expires от SDS.
type orderExpiresPayload struct {
	ID          string
	Status      string
	EndTime     string
	ProductCode string
	Phone       string
	Lang        string
}

// Сценарий 3: СМС №5 за день до окончания триала/подписки.
// Шаг 1 (SDS): периодическая проверка → публикация order.expires в NATS.
// Шаги 2–4 (viva-api): приём → продукт по externalId → СМС №5 по SMPP.

func (s *viva) orderExpiresHandler(ctx types.HandlerContext) {
	payload, err := parseOrderExpiresEvent(ctx)
	if err != nil {
		logger.Error().Err(err).Str("topic", orderExpiresSubscribePath).Msg("scenario 3 step 2: failed to parse order.expires")
		return
	}

	scenario3Step2Received(payload)

	if err := s.handleOrderExpires(payload); err != nil {
		logger.Error().Err(err).Str("orderId", payload.ID).Msg("scenario 3: handle order.expires failed")
	}
}

func scenario3Step2Received(p orderExpiresPayload) {
	logger.Info().
		Str("orderId", p.ID).
		Str("status", p.Status).
		Str("endTime", p.EndTime).
		Str("natsSubject", "order.expires").
		Msg("scenario 3 step 2: viva-service received order.expires from NATS")
}

func parseOrderExpiresEvent(ctx types.HandlerContext) (orderExpiresPayload, error) {
	var expire types.OrderExpireRequest
	ctx.Data(&expire)
	if id := strings.TrimSpace(expire.ID); id != "" {
		return orderExpiresPayload{
			ID:      id,
			Status:  strings.TrimSpace(expire.Status),
			EndTime: strings.TrimSpace(expire.EndTime),
		}, nil
	}

	var order types.OrderResponse
	ctx.Data(&order)
	if id := strings.TrimSpace(order.ID); id != "" {
		return orderExpiresPayloadFromOrder(order), nil
	}

	return orderExpiresPayload{}, fmt.Errorf("order.expires: empty order id in payload")
}

func orderExpiresPayloadFromOrder(order types.OrderResponse) orderExpiresPayload {
	endTime := ""
	if order.EndTime != nil {
		endTime = order.EndTime.UTC().Format(time.RFC3339)
	}
	lang := ""
	if order.CustomData != nil {
		lang = strings.TrimSpace(cast.ToString(order.CustomData["lang"]))
	}
	return orderExpiresPayload{
		ID:          strings.TrimSpace(order.ID),
		Status:      strings.TrimSpace(order.Status),
		EndTime:     endTime,
		ProductCode: productCodeFromOrderResponse(order),
		Phone:       phoneFromOrderResponse(order),
		Lang:        lang,
	}
}

// ParseOrderExpiresJSON разбирает сырое тело сообщения NATS (тесты и утилиты).
func ParseOrderExpiresJSON(raw []byte) (orderExpiresPayload, error) {
	var expire types.OrderExpireRequest
	if err := json.Unmarshal(raw, &expire); err == nil && strings.TrimSpace(expire.ID) != "" {
		return orderExpiresPayload{
			ID:      expire.ID,
			Status:  expire.Status,
			EndTime: expire.EndTime,
		}, nil
	}
	var order types.OrderResponse
	if err := json.Unmarshal(raw, &order); err != nil {
		return orderExpiresPayload{}, err
	}
	if strings.TrimSpace(order.ID) == "" {
		return orderExpiresPayload{}, fmt.Errorf("order.expires: empty order id")
	}
	return orderExpiresPayloadFromOrder(order), nil
}

func productCodeFromOrderResponse(order types.OrderResponse) string {
	if order.CustomData != nil {
		if c := strings.TrimSpace(cast.ToString(order.CustomData["productCode"])); c != "" {
			return c
		}
	}
	if len(order.Items) > 0 {
		if c := strings.TrimSpace(order.Items[0].Product.Name); c != "" {
			return c
		}
	}
	return ""
}

func phoneFromOrderResponse(order types.OrderResponse) string {
	if order.Fields == nil {
		return ""
	}
	return strings.TrimSpace(cast.ToString(order.Fields["phone"]))
}

func (s *viva) enrichOrderExpiresPayload(p orderExpiresPayload) (orderExpiresPayload, error) {
	if strings.TrimSpace(p.ProductCode) != "" {
		return p, nil
	}
	order, err := s.getOrder(p.ID)
	if err != nil {
		return p, fmt.Errorf("order/get for order.expires: %w", err)
	}
	enriched := orderExpiresPayloadFromOrder(order)
	if enriched.ProductCode == "" {
		return p, fmt.Errorf("order.expires: productCode missing in order %s", p.ID)
	}
	if p.Phone == "" {
		p.Phone = enriched.Phone
	}
	p.ProductCode = enriched.ProductCode
	if p.Lang == "" {
		p.Lang = enriched.Lang
	}
	if p.Status == "" {
		p.Status = enriched.Status
	}
	if p.EndTime == "" {
		p.EndTime = enriched.EndTime
	}
	return p, nil
}

func formatExpiresAtForSMS(endTime string) string {
	endTime = strings.TrimSpace(endTime)
	if endTime == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, endTime); err == nil {
			return t.Format("02.01.2006 15:04")
		}
	}
	return endTime
}

// scenario3Step4 отправляет СМС №5 (напоминание об окончании триала) по SMPP.
func (s *viva) scenario3Step4(p orderExpiresPayload, product catalog.Product) error {
	phone := normalizePhone(p.Phone)
	if phone == "" {
		return fmt.Errorf("order.expires: phone is empty")
	}

	lang := s.resolveLang(phone, p.Lang)
	if lang == "" {
		lang = normalizeLang(product.GetDefaultLanguage())
	}

	expiresAt := formatExpiresAtForSMS(p.EndTime)
	data := map[string]interface{}{
		"Phone":       phone,
		"ExternalID":  p.ProductCode,
		"ExpiresAt":   expiresAt,
		"EndTime":     p.EndTime,
		"Language":    lang,
	}

	text := product.GetNotify(NotifyTrialExpires, data, lang)
	if text == "" {
		return fmt.Errorf("trial_expires notification template is empty")
	}

	logger.Info().
		Str("phone", phone).
		Str("lang", lang).
		Str("expiresAt", expiresAt).
		Msg("scenario 3 step 4: sending SMS #5 (trial expires) via SMPP")

	if err := s.notify(phone, text); err != nil {
		return fmt.Errorf("send trial_expires sms: %w", err)
	}
	return nil
}

func (s *viva) handleOrderExpires(p orderExpiresPayload) error {
	if strings.TrimSpace(p.ID) == "" {
		return errOrderExpiresNoID{}
	}

	p, err := s.enrichOrderExpiresPayload(p)
	if err != nil {
		return err
	}

	product, err := s.productForExternalID(FlowTrial, "3", p.ProductCode)
	if err != nil {
		return err
	}

	return s.scenario3Step4(p, product)
}

type errOrderExpiresNoID struct{}

func (errOrderExpiresNoID) Error() string { return "order.expires: order id is empty" }
