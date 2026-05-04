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
	// Locale: en | ru | hy — из customData заказа (лендинг / вебхук).
	Locale string
}

func BuildSmsText(sc SmsScenario, p SmsScenarioPayload) string {
	lang := LocaleOrDefault(p.Locale)
	product := strings.TrimSpace(p.ProductLabel)
	if product == "" {
		product = defaultProductNameI18n(lang)
	}
	switch sc {
	case Sms2Welcome:
		return sms2Welcome(lang, product)
	case Sms3License:
		return sms3License(lang, product, p.LicenseKey, p.DownloadURL)
	case Sms4Paid:
		return sms4Paid(lang, product)
	case Sms5TrialRemind:
		return sms5Trial(lang, product, p.TrialEndDate)
	case Sms14NoFunds:
		return sms14NoFunds(lang)
	case Sms15Booster:
		return sms15Booster(lang)
	case SmsServiceRemoved:
		return smsDeactivated(lang, product)
	default:
		return ""
	}
}

func defaultProductNameI18n(lang string) string {
	switch lang {
	case SmsLangEN:
		return "Kaspersky Safe Kids"
	case SmsLangHY:
		return "Kaspersky Safe Kids"
	default:
		return "Kaspersky Safe Kids"
	}
}

func sms2Welcome(lang, product string) string {
	switch lang {
	case SmsLangEN:
		return fmt.Sprintf("Congratulations! %s is active. 2-day trial, then 50 AMD/day.", product)
	case SmsLangHY:
		return fmt.Sprintf("Շնորհավորում ենք: %s ակտիվ է: 2 օր փորձ, այնուհետև 50 դրամ/օր:", product)
	default:
		return fmt.Sprintf("Поздравляем! Подписка %s активна. Пробный период 2 дня. Далее 50 AMD/день.", product)
	}
}

func sms3License(lang, product, key, url string) string {
	key, url = strings.TrimSpace(key), strings.TrimSpace(url)
	switch lang {
	case SmsLangEN:
		return fmt.Sprintf(`"%s": %s. Link: %s`, product, key, url)
	case SmsLangHY:
		return fmt.Sprintf(`"%s": %s. Հղում: %s`, product, key, url)
	default:
		return fmt.Sprintf(`"%s": %s. Ссылка: %s`, product, key, url)
	}
}

func sms4Paid(lang, product string) string {
	switch lang {
	case SmsLangEN:
		return fmt.Sprintf("Chargeable %s is active. 50 AMD/day.", product)
	case SmsLangHY:
		return fmt.Sprintf("%s վճարովի տարբերակը միացված է: 50 դրամ/օր:", product)
	default:
		return fmt.Sprintf("Платная версия %s подключена. Цена 50 AMD/день.", product)
	}
}

func sms5Trial(lang, product, trialEnd string) string {
	d := strings.TrimSpace(trialEnd)
	switch lang {
	case SmsLangEN:
		if d != "" {
			return fmt.Sprintf("%s trial ends %s. Paid plan will renew if balance is OK.", product, d)
		}
		return fmt.Sprintf("%s trial ends soon.", product)
	case SmsLangHY:
		if d != "" {
			return fmt.Sprintf("%s փորձնական շրջանը ավարտվում է %s:", product, d)
		}
		return fmt.Sprintf("%s փորձնական շրջանը շուտով ավարտվում է:", product)
	default:
		if d != "" {
			return fmt.Sprintf("Пробный период %s истекает %s. Пополните счёт вовремя.", product, d)
		}
		return fmt.Sprintf("Пробный период %s скоро закончится.", product)
	}
}

func sms14NoFunds(lang string) string {
	switch lang {
	case SmsLangEN:
		return `Top up via *221# or https://cabinet.viva.am/epay2/ to continue Kaspersky Safe Kids.`
	case SmsLangHY:
		return `Լիցքավորեք հաշիվը *221# կամ https://cabinet.viva.am/epay2/ Kaspersky Safe Kids-ի համար:`
	default:
		return "Пополните счёт через *221# или https://cabinet.viva.am/epay2/ для продолжения Kaspersky Safe Kids."
	}
}

func sms15Booster(lang string) string {
	switch lang {
	case SmsLangEN:
		return "Long time no see. Top up your balance to continue Kaspersky Safe Kids."
	case SmsLangHY:
		return "Վաղուց չենք հանդիպել: Լիցքավորեք հաշիվը՝ շարունակելու Kaspersky Safe Kids:"
	default:
		return "Давно не встречались. Пополните баланс для продолжения Kaspersky Safe Kids."
	}
}

func smsDeactivated(lang, product string) string {
	switch lang {
	case SmsLangEN:
		return fmt.Sprintf("%s service deactivated.", product)
	case SmsLangHY:
		return fmt.Sprintf("%s ծառայությունն ապաակտիվացված է:", product)
	default:
		return fmt.Sprintf("Услуга %s деактивирована.", product)
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
	logger.Info().Str("sender", "SMPP").Str("scenario", string(sc)).Str("msisdn", msisdn).Msg("SPEC-8 SMS")
	return s.smppSender.Send(msisdn, body)
}
