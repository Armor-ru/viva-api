package viva

import (
	"os"
	"strings"
	"testing"
	"text/template"

	"gopkg.in/yaml.v3"
)

func TestSMSTemplatesMatchLegacyTexts(t *testing.T) {
	product := "Kaspersky Safe Kids"
	templates := loadSMSTemplates(t)

	cases := []struct {
		scenario string
		lang     string
		data     map[string]string
		want     string
	}{
		{"sms2", "en", map[string]string{"ProductLabel": product}, "Congratulations! Kaspersky Safe Kids is active. 2-day trial, then 50 AMD/day."},
		{"sms2", "ru", map[string]string{"ProductLabel": product}, "Поздравляем! Подписка Kaspersky Safe Kids активна. Пробный период 2 дня. Далее 50 AMD/день."},
		{"sms2", "hy", map[string]string{"ProductLabel": product}, "Շնորհավորում ենք: Kaspersky Safe Kids ակտիվ է: 2 օր փորձ, այնուհետև 50 դրամ/օր:"},
		{"sms3", "en", map[string]string{"ProductLabel": product, "LicenseKey": "KEY1", "DownloadURL": "https://dl.example/x"}, `"Kaspersky Safe Kids": KEY1. Link: https://dl.example/x`},
		{"sms3", "ru", map[string]string{"ProductLabel": product, "LicenseKey": "KEY1", "DownloadURL": "https://dl.example/x"}, `"Kaspersky Safe Kids": KEY1. Ссылка: https://dl.example/x`},
		{"sms3", "hy", map[string]string{"ProductLabel": product, "LicenseKey": "KEY1", "DownloadURL": "https://dl.example/x"}, `"Kaspersky Safe Kids": KEY1. Հղում: https://dl.example/x`},
		{"sms4", "en", map[string]string{"ProductLabel": product}, "Chargeable Kaspersky Safe Kids is active. 50 AMD/day."},
		{"sms4", "ru", map[string]string{"ProductLabel": product}, "Платная версия Kaspersky Safe Kids подключена. Цена 50 AMD/день."},
		{"sms4", "hy", map[string]string{"ProductLabel": product}, "Kaspersky Safe Kids վճարովի տարբերակը միացված է: 50 դրամ/օր:"},
		{"sms5", "en", map[string]string{"ProductLabel": product, "TrialEndDate": "18.05.2026"}, "Kaspersky Safe Kids trial ends 18.05.2026. Paid plan will renew if balance is OK."},
		{"sms5", "en", map[string]string{"ProductLabel": product}, "Kaspersky Safe Kids trial ends soon."},
		{"sms5", "ru", map[string]string{"ProductLabel": product, "TrialEndDate": "18.05.2026"}, "Пробный период Kaspersky Safe Kids истекает 18.05.2026. Пополните счёт вовремя."},
		{"sms5", "ru", map[string]string{"ProductLabel": product}, "Пробный период Kaspersky Safe Kids скоро закончится."},
		{"sms5", "hy", map[string]string{"ProductLabel": product, "TrialEndDate": "18.05.2026"}, "Kaspersky Safe Kids փորձնական շրջանը ավարտվում է 18.05.2026:"},
		{"sms5", "hy", map[string]string{"ProductLabel": product}, "Kaspersky Safe Kids փորձնական շրջանը շուտով ավարտվում է:"},
		{"sms14", "en", map[string]string{"ProductLabel": product}, "Top up via *221# or https://viva.am to continue Kaspersky Safe Kids."},
		{"sms14", "ru", map[string]string{"ProductLabel": product}, "Пополните счёт через *221# или https://viva.am для продолжения Kaspersky Safe Kids."},
		{"sms14", "hy", map[string]string{"ProductLabel": product}, "Լիցքավորեք հաշիվը *221# կամ https://viva.am Kaspersky Safe Kids-ի համար:"},
		{"sms15", "en", map[string]string{"ProductLabel": product}, "Long time no see. Top up your balance to continue Kaspersky Safe Kids."},
		{"sms15", "ru", map[string]string{"ProductLabel": product}, "Давно не встречались. Пополните баланс для продолжения Kaspersky Safe Kids."},
		{"sms15", "hy", map[string]string{"ProductLabel": product}, "Վաղուց չենք հանդիպել: Լիցքավորեք հաշիվը՝ շարունակելու Kaspersky Safe Kids:"},
		{"sms_deactivated", "en", map[string]string{"ProductLabel": product}, "Kaspersky Safe Kids service deactivated."},
		{"sms_deactivated", "ru", map[string]string{"ProductLabel": product}, "Услуга Kaspersky Safe Kids деактивирована."},
		{"sms_deactivated", "hy", map[string]string{"ProductLabel": product}, "Kaspersky Safe Kids ծառայությունն ապաակտիվացված է:"},
	}

	for _, tc := range cases {
		t.Run(tc.scenario+"_"+tc.lang, func(t *testing.T) {
			text := templates[tc.scenario][tc.lang]
			tmpl, err := template.New(tc.scenario + ":" + tc.lang).Parse(text)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			var got strings.Builder
			if err := tmpl.Execute(&got, tc.data); err != nil {
				t.Fatalf("execute: %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("got %q\nwant %q", got.String(), tc.want)
			}
		})
	}
}

func loadSMSTemplates(t *testing.T) map[string]map[string]string {
	t.Helper()
	raw, err := os.ReadFile("../../config/viva-api.yaml")
	if err != nil {
		t.Fatalf("read templates: %v", err)
	}
	var cfg struct {
		SMSTemplates map[string]map[string]string `yaml:"smsTemplates"`
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse templates: %v", err)
	}
	return cfg.SMSTemplates
}
