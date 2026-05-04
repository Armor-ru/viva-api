package viva_api

import (
	"strings"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/spf13/cast"
)

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
	logger.Info().Str("orderId", order.ID).Str("phone", phone).Msg("SPEC-8 sms2+sms3 sent")
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
