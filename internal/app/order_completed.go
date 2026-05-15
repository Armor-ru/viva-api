package viva_api

import (
	"strings"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/spf13/cast"
)

func (s *Viva) onOrderCompleted(ctx types.HandlerContext) {
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
	locale := localeOrDefault(cast.ToString(order.CustomData[CDSmsLocale]))

	switch wh {
	case WHActivation:
		s.sendRenewSMS(order, phone, scenarioHint, locale)
	case WHRemove:
		s.sendRemoveSMS(order, phone, scenarioHint, locale)
	default:
		s.sendNewActivationSMS(order, phone, locale)
	}
}

func orderPhone(order types.OrderResponse) string {
	raw, ok := order.Fields["phone"]
	if !ok {
		return ""
	}
	return strings.TrimSpace(cast.ToString(raw))
}

func (s *Viva) sendNewActivationSMS(order types.OrderResponse, phone, locale string) {
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

	label := cast.ToString(item.Product.Name)
	payload := SmsTplData{
		ProductName:    label,
		ProductLabel:   label,
		ActivationCode: activationCode,
		LicenseKey:     activationCode,
		DownloadURL:    downloadURL,
	}
	if quantity, ok := item.Options["quantity"]; ok {
		payload.Quantity = cast.ToInt(quantity)
	}

	for _, sc := range []string{scSms2, scSms3} {
		if err := s.sendScenario(phone, sc, locale, payload); err != nil {
			logger.Error().Str("orderId", order.ID).Err(err).Str("scenario", sc).Msg("SMPP")
			return
		}
	}
	logger.Info().Str("orderId", order.ID).Str("phone", phone).Msg("activation sms sent")
}

func (s *Viva) sendRenewSMS(order types.OrderResponse, phone, scenarioHint, locale string) {
	sc := scSms4
	switch scenarioHint {
	case scSms5, scSms15, scSms4:
		sc = scenarioHint
	}
	payload := SmsTplData{ProductLabel: productLabel(order)}
	if sc == scSms5 && order.EndTime != nil {
		payload.TrialEndDate = order.EndTime.Format("02.01.2006")
	}
	if err := s.sendScenario(phone, sc, locale, payload); err != nil {
		logger.Error().Str("orderId", order.ID).Err(err).Str("scenario", sc).Msg("SMPP renew")
	}
}

func (s *Viva) sendRemoveSMS(order types.OrderResponse, phone, scenarioHint, locale string) {
	sc := scOff
	if scenarioHint == scSms14 {
		sc = scSms14
	}
	payload := SmsTplData{ProductLabel: productLabel(order)}
	if err := s.sendScenario(phone, sc, locale, payload); err != nil {
		logger.Error().Str("orderId", order.ID).Err(err).Str("scenario", sc).Msg("SMPP remove")
	}
}

func productLabel(order types.OrderResponse) string {
	if len(order.Items) == 0 {
		return ""
	}
	return cast.ToString(order.Items[0].Product.Name)
}
