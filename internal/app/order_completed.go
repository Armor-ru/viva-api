package viva_api

import (
	"strings"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/spf13/cast"
)

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
