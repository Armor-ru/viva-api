package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

func (s *viva) notify(phone, text string) error {
	phone = normalizePhone(phone)
	text = strings.TrimSpace(text)
	if phone == "" || text == "" {
		return fmt.Errorf("sms phone and text are required")
	}
	if s.ussdTransport == nil {
		return fmt.Errorf("ussd transport is not configured")
	}

	_, err := s.ussdTransport.Send("", map[string]interface{}{
		"to":   phone,
		"text": text,
	}, types.SendOptions{})
	if err != nil {
		return err
	}
	logger.Debug().Str("phone", phone).Int("len", len(text)).Msg("smpp: outbound SMS submitted")
	return nil
}

func (s *viva) orderCompleteHandler(ctx types.HandlerContext) {
	order := types.OrderResponse{}
	ctx.Data(&order)

	if isScenario4OrderCompleted(order) {
		if err := s.handleScenario4OrderCompleted(order); err != nil {
			logger.Error().Err(err).Str("orderId", order.ID).Msg("scenario 4: handle order/completed failed")
		}
		return
	}

	logger.Info().
		Str("orderId", order.ID).
		Str("status", order.Status).
		Str("topic", "order/completed").
		Msg("scenario step 10: order/completed delivered from NATS (published by SDS)")

	if err := s.completeOrder(order); err != nil {
		logger.Error().Err(err).Str("orderId", order.ID).Msg("scenario steps 11–13: processing order/completed failed")
		return
	}
	flowInfo(FlowOrder, "completed").Str("orderId", order.ID).Msg("order/completed handled, outbound SMS sent if configured")
}

// sendActivationSMS отправляет СМС №2 (приветствие, триал) и №3 (лицензия) после order/completed.
func (s *viva) sendActivationSMS(phone, lang string, product catalog.Product, data map[string]interface{}) error {
	if text := product.GetNotify(NotifyWelcomeTrial, data, lang); text != "" {
		flowInfo(FlowOrder, "completed").Str("phone", phone).Str("lang", lang).Msg("sending SMS #2 via SMPP")
		if err := s.notify(phone, text); err != nil {
			return fmt.Errorf("send welcome sms: %w", err)
		}
	}
	if text := product.GetNotify(NotifyLicense, data, lang); text != "" {
		flowInfo(FlowOrder, "completed").Str("phone", phone).Str("lang", lang).Msg("sending SMS #3 via SMPP")
		if err := s.notify(phone, text); err != nil {
			return fmt.Errorf("send license sms: %w", err)
		}
	}
	return nil
}
