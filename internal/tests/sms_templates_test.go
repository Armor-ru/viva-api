package viva

import (
	"fmt"
	"testing"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
)

func TestSMSTemplatesMatchLegacyTexts(t *testing.T) {
	product := "Kaspersky Safe Kids"

	cases := []struct {
		scenario string
		lang     string
		args     []interface{}
		want     string
	}{
		{"sms2", "en", []interface{}{product}, "Congratulations! Kaspersky Safe Kids is active. 2-day trial, then 50 AMD/day."},
		{"sms2", "ru", []interface{}{product}, "Поздравляем! Подписка Kaspersky Safe Kids активна. Пробный период 2 дня. Далее 50 AMD/день."},
		{"sms2", "hy", []interface{}{product}, "Շնորհավորում ենք: Kaspersky Safe Kids ակտիվ է: 2 օր փորձ, այնուհետև 50 դրամ/օր:"},
		{"sms3", "en", []interface{}{product, "KEY1", "https://dl.example/x"}, `"Kaspersky Safe Kids": KEY1. Link: https://dl.example/x`},
		{"sms3", "ru", []interface{}{product, "KEY1", "https://dl.example/x"}, `"Kaspersky Safe Kids": KEY1. Ссылка: https://dl.example/x`},
		{"sms3", "hy", []interface{}{product, "KEY1", "https://dl.example/x"}, `"Kaspersky Safe Kids": KEY1. Հղում: https://dl.example/x`},
		{"sms4", "en", []interface{}{product}, "Chargeable Kaspersky Safe Kids is active. 50 AMD/day."},
		{"sms4", "ru", []interface{}{product}, "Платная версия Kaspersky Safe Kids подключена. Цена 50 AMD/день."},
		{"sms4", "hy", []interface{}{product}, "Kaspersky Safe Kids վճարովի տարբերակը միացված է: 50 դրամ/օր:"},
		{"sms5_with_date", "en", []interface{}{product, "18.05.2026"}, "Kaspersky Safe Kids trial ends 18.05.2026. Paid plan will renew if balance is OK."},
		{"sms5_soon", "en", []interface{}{product}, "Kaspersky Safe Kids trial ends soon."},
		{"sms5_with_date", "ru", []interface{}{product, "18.05.2026"}, "Пробный период Kaspersky Safe Kids истекает 18.05.2026. Пополните счёт вовремя."},
		{"sms5_soon", "ru", []interface{}{product}, "Пробный период Kaspersky Safe Kids скоро закончится."},
		{"sms5_with_date", "hy", []interface{}{product, "18.05.2026"}, "Kaspersky Safe Kids փորձնական շրջանը ավարտվում է 18.05.2026:"},
		{"sms5_soon", "hy", []interface{}{product}, "Kaspersky Safe Kids փորձնական շրջանը շուտով ավարտվում է:"},
		{"sms14", "en", nil, "Top up via *221# or https://viva.am to continue Kaspersky Safe Kids."},
		{"sms14", "ru", nil, "Пополните счёт через *221# или https://viva.am для продолжения Kaspersky Safe Kids."},
		{"sms14", "hy", nil, "Լիցքավորեք հաշիվը *221# կամ https://viva.am Kaspersky Safe Kids-ի համար:"},
		{"sms15", "en", nil, "Long time no see. Top up your balance to continue Kaspersky Safe Kids."},
		{"sms15", "ru", nil, "Давно не встречались. Пополните баланс для продолжения Kaspersky Safe Kids."},
		{"sms15", "hy", nil, "Վաղուց չենք հանդիպել: Լիցքավորեք հաշիվը՝ շարունակելու Kaspersky Safe Kids:"},
		{"sms_deactivated", "en", []interface{}{product}, "Kaspersky Safe Kids service deactivated."},
		{"sms_deactivated", "ru", []interface{}{product}, "Услуга Kaspersky Safe Kids деактивирована."},
		{"sms_deactivated", "hy", []interface{}{product}, "Kaspersky Safe Kids ծառայությունն ապաակտիվացված է:"},
	}

	for _, tc := range cases {
		t.Run(tc.scenario+"_"+tc.lang, func(t *testing.T) {
			got := fmt.Sprintf(viva_api.GetTemplate(tc.scenario, tc.lang), tc.args...)
			if got != tc.want {
				t.Fatalf("got %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestGetTemplateFallback(t *testing.T) {
	got := viva_api.GetTemplate("unknown", "en")
	if got != "" {
		t.Fatalf("got %q", got)
	}
}
