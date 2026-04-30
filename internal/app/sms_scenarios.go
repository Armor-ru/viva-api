package viva_api

import (
	"fmt"
	"strings"

	"github.com/Armor-ru/sds-go/pkg/logger"
)

type SmsScenario string

const (
	Sms2Welcome       SmsScenario = "sms2"
	Sms3License       SmsScenario = "sms3"
	Sms4Paid          SmsScenario = "sms4"
	Sms5TrialRemind   SmsScenario = "sms5"
	Sms14NoFunds      SmsScenario = "sms14"
	Sms15Booster      SmsScenario = "sms15"
	SmsServiceRemoved SmsScenario = "sms_deactivated"
)

type SmsScenarioPayload struct {
	TrialEndDate string
	LicenseKey   string
	DownloadURL  string
	ProductLabel string
}

func BuildSmsText(sc SmsScenario, p SmsScenarioPayload) string {
	product := strings.TrimSpace(p.ProductLabel)
	if product == "" {
		product = "Kaspersky Safe Kids"
	}
	switch sc {
	case Sms2Welcome:
		return fmt.Sprintf("Поздравляем! Подписка %s активна. Пробный период 2 дня. Далее 50 AMD/день.", product)
	case Sms3License:
		return fmt.Sprintf("%s: %s. Ссылка: %s", product, strings.TrimSpace(p.LicenseKey), strings.TrimSpace(p.DownloadURL))
	case Sms4Paid:
		return fmt.Sprintf("Платная версия %s подключена. Цена 50 AMD/день.", product)
	case Sms5TrialRemind:
		d := strings.TrimSpace(p.TrialEndDate)
		if d != "" {
			return fmt.Sprintf("Пробный период %s истекает %s. Пополните счёт вовремя.", product, d)
		}
		return fmt.Sprintf("Пробный период %s скоро закончится.", product)
	case Sms14NoFunds:
		return "Пополните счёт через *221# или https://cabinet.viva.am/epay2/ для продолжения Kaspersky Safe Kids."
	case Sms15Booster:
		return "Давно не встречались. Пополните баланс для продолжения Kaspersky Safe Kids."
	case SmsServiceRemoved:
		return fmt.Sprintf("Услуга %s деактивирована.", product)
	default:
		return ""
	}
}

func (s *Viva) smppSendScenario(msisdn string, sc SmsScenario, p SmsScenarioPayload) error {
	if s.smppSender == nil {
		return fmt.Errorf("smppSender is nil")
	}
	body := BuildSmsText(sc, p)
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("empty sms body for scenario %s", sc)
	}
	logger.Info().Str("sender", "SMPP").Str("scenario", string(sc)).Str("msisdn", msisdn).Msg("SPEC §8 SMS")
	return s.smppSender.Send(msisdn, body)
}
