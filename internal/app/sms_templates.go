package viva_api

var smsTemplates = map[string]string{
	"sms2:en": "Congratulations! %s is active. 2-day trial, then 50 AMD/day.",
	"sms2:ru": "Поздравляем! Подписка %s активна. Пробный период 2 дня. Далее 50 AMD/день.",
	"sms2:hy": "Շնորհավորում ենք: %s ակտիվ է: 2 օր փորձ, այնուհետև 50 դրամ/օր:",

	"sms3:en": `"%s": %s. Link: %s`,
	"sms3:ru": `"%s": %s. Ссылка: %s`,
	"sms3:hy": `"%s": %s. Հղում: %s`,

	"sms4:en": "Chargeable %s is active. 50 AMD/day.",
	"sms4:ru": "Платная версия %s подключена. Цена 50 AMD/день.",
	"sms4:hy": "%s վճարովի տարբերակը միացված է: 50 դրամ/օր:",

	"sms5_with_date:en": "%s trial ends %s. Paid plan will renew if balance is OK.",
	"sms5_with_date:ru": "Пробный период %s истекает %s. Пополните счёт вовремя.",
	"sms5_with_date:hy": "%s փորձնական շրջանը ավարտվում է %s:",

	"sms5_soon:en": "%s trial ends soon.",
	"sms5_soon:ru": "Пробный период %s скоро закончится.",
	"sms5_soon:hy": "%s փորձնական շրջանը շուտով ավարտվում է:",

	"sms14:en": "Top up via *221# or https://viva.am to continue Kaspersky Safe Kids.",
	"sms14:ru": "Пополните счёт через *221# или https://viva.am для продолжения Kaspersky Safe Kids.",
	"sms14:hy": "Լիցքավորեք հաշիվը *221# կամ https://viva.am Kaspersky Safe Kids-ի համար:",

	"sms15:en": "Long time no see. Top up your balance to continue Kaspersky Safe Kids.",
	"sms15:ru": "Давно не встречались. Пополните баланс для продолжения Kaspersky Safe Kids.",
	"sms15:hy": "Վաղուց չենք հանդիպել: Լիցքավորեք հաշիվը՝ շարունակելու Kaspersky Safe Kids:",

	"sms_deactivated:en": "%s service deactivated.",
	"sms_deactivated:ru": "Услуга %s деактивирована.",
	"sms_deactivated:hy": "%s ծառայությունն ապաակտիվացված է:",
}

func GetTemplate(scenario, lang string) string {
	lang = localeOrDefault(lang)
	if text, ok := smsTemplates[scenario+":"+lang]; ok {
		return text
	}
	return smsTemplates[scenario+":en"]
}
