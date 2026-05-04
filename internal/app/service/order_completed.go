package service

import (
	"strings"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
	"github.com/Armor-ru/viva-api/internal/app/utils"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/spf13/cast"
)

func HandleOrderCompleted(v *viva_api.Viva, ctx types.HandlerContext) {
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

	wh := strings.TrimSpace(cast.ToString(order.CustomData[viva_api.CDVivaWebhook]))
	scenarioHint := strings.TrimSpace(cast.ToString(order.CustomData[viva_api.CDSmsScenario]))
	locale := orderSmsLocale(order)

	switch wh {
	case viva_api.WHActivation:
		sendRenewSms(v, order, phone, scenarioHint, locale)
	case viva_api.WHRemove:
		sendRemoveSms(v, order, phone, scenarioHint, locale)
	default:
		sendNewActivationSms(v, order, phone, locale)
	}
}

func orderSmsLocale(order types.OrderResponse) string {
	return utils.LocaleOrDefault(cast.ToString(order.CustomData[viva_api.CDSmsLocale]))
}

func orderPhone(order types.OrderResponse) string {
	raw, ok := order.Fields["phone"]
	if !ok {
		return ""
	}
	return strings.TrimSpace(cast.ToString(raw))
}
