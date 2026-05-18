package viva_api

import (
	"fmt"
	"strings"
)

func smsText(scenario, lang string, d SmsData) string {
	product := strings.TrimSpace(d.ProductLabel)
	if product == "" {
		product = "Kaspersky Safe Kids"
	}
	lang = localeOrDefault(lang)

	switch scenario {
	case "sms2":
		switch lang {
		case "en":
			return fmt.Sprintf("Congratulations! %s is active. 2-day trial, then 50 AMD/day.", product)
		case "hy":
			return fmt.Sprintf("Շնորհավորում ենք: %s ակտիվ է: 2 օր փորձ, այնուհետև 50 դրամ/օր:", product)
		default:
			return fmt.Sprintf("Поздравляем! Подписка %s активна. Пробный период 2 дня. Далее 50 AMD/день.", product)
		}
	case "sms3":
		key, url := strings.TrimSpace(d.LicenseKey), strings.TrimSpace(d.DownloadURL)
		switch lang {
		case "en":
			return fmt.Sprintf(`"%s": %s. Link: %s`, product, key, url)
		case "hy":
			return fmt.Sprintf(`"%s": %s. Հղում: %s`, product, key, url)
		default:
			return fmt.Sprintf(`"%s": %s. Ссылка: %s`, product, key, url)
		}
	case "sms4":
		switch lang {
		case "en":
			return fmt.Sprintf("Chargeable %s is active. 50 AMD/day.", product)
		case "hy":
			return fmt.Sprintf("%s վճարովի տարբերակը միացված է: 50 դրամ/օր:", product)
		default:
			return fmt.Sprintf("Платная версия %s подключена. Цена 50 AMD/день.", product)
		}
	case "sms5":
		end := strings.TrimSpace(d.TrialEndDate)
		switch lang {
		case "en":
			if end != "" {
				return fmt.Sprintf("%s trial ends %s. Paid plan will renew if balance is OK.", product, end)
			}
			return fmt.Sprintf("%s trial ends soon.", product)
		case "hy":
			if end != "" {
				return fmt.Sprintf("%s փորձնական շրջանը ավարտվում է %s:", product, end)
			}
			return fmt.Sprintf("%s փորձնական շրջանը շուտով ավարտվում է:", product)
		default:
			if end != "" {
				return fmt.Sprintf("Пробный период %s истекает %s. Пополните счёт вовремя.", product, end)
			}
			return fmt.Sprintf("Пробный период %s скоро закончится.", product)
		}
	case "sms14":
		switch lang {
		case "en":
			return `Top up via *221# or https://cabinet.viva.am/epay2/ to continue Kaspersky Safe Kids.`
		case "hy":
			return `Լիցքավորեք հաշիվը *221# կամ https://cabinet.viva.am/epay2/ Kaspersky Safe Kids-ի համար:`
		default:
			return "Пополните счёт через *221# или https://cabinet.viva.am/epay2/ для продолжения Kaspersky Safe Kids."
		}
	case "sms15":
		switch lang {
		case "en":
			return "Long time no see. Top up your balance to continue Kaspersky Safe Kids."
		case "hy":
			return "Վաղուց չենք հանդիպել: Լիցքավորեք հաշիվը՝ շարունակելու Kaspersky Safe Kids:"
		default:
			return "Давно не встречались. Пополните баланс для продолжения Kaspersky Safe Kids."
		}
	case "sms_deactivated":
		switch lang {
		case "en":
			return fmt.Sprintf("%s service deactivated.", product)
		case "hy":
			return fmt.Sprintf("%s ծառայությունն ապաակտիվացված է:", product)
		default:
			return fmt.Sprintf("Услуга %s деактивирована.", product)
		}
	default:
		return ""
	}
}
