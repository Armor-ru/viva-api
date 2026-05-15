package viva_api

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/tplext"
)

func defaultSmsTemplates() map[string]map[string]string {
	return map[string]map[string]string{
		scSms2: {
			smsLangEN: "Congratulations! {{.ProductLabel}} is active. 2-day trial, then 50 AMD/day.",
			smsLangHY: "Շնորհավորում ենք: {{.ProductLabel}} ակտիվ է: 2 օր փորձ, այնուհետև 50 դրամ/օր:",
			smsLangRU: "Поздравляем! Подписка {{.ProductLabel}} активна. Пробный период 2 дня. Далее 50 AMD/день.",
		},
		scSms3: {
			smsLangEN: `"{{.ProductLabel}}": {{.LicenseKey}}. Link: {{.DownloadURL}}`,
			smsLangHY: `"{{.ProductLabel}}": {{.LicenseKey}}. Հղում: {{.DownloadURL}}`,
			smsLangRU: `"{{.ProductLabel}}": {{.LicenseKey}}. Ссылка: {{.DownloadURL}}`,
		},
		scSms4: {
			smsLangEN: "Chargeable {{.ProductLabel}} is active. 50 AMD/day.",
			smsLangHY: "{{.ProductLabel}} վճարովի տարբերակը միացված է: 50 դրամ/օր:",
			smsLangRU: "Платная версия {{.ProductLabel}} подключена. Цена 50 AMD/день.",
		},
		scSms5: {
			smsLangEN: `{{if .TrialEndDate}}{{.ProductLabel}} trial ends {{.TrialEndDate}}. Paid plan will renew if balance is OK.{{else}}{{.ProductLabel}} trial ends soon.{{end}}`,
			smsLangHY: `{{if .TrialEndDate}}{{.ProductLabel}} փորձնական շրջանը ավարտվում է {{.TrialEndDate}}:{{else}}{{.ProductLabel}} փորձնական շրջանը շուտով ավարտվում է:{{end}}`,
			smsLangRU: `{{if .TrialEndDate}}Пробный период {{.ProductLabel}} истекает {{.TrialEndDate}}. Пополните счёт вовремя.{{else}}Пробный период {{.ProductLabel}} скоро закончится.{{end}}`,
		},
		scSms14: {
			smsLangEN: `Top up via *221# or https://cabinet.viva.am/epay2/ to continue Kaspersky Safe Kids.`,
			smsLangHY: `Լիցքավորեք հաշիվը *221# կամ https://cabinet.viva.am/epay2/ Kaspersky Safe Kids-ի համար:`,
			smsLangRU: "Пополните счёт через *221# или https://cabinet.viva.am/epay2/ для продолжения Kaspersky Safe Kids.",
		},
		scSms15: {
			smsLangEN: "Long time no see. Top up your balance to continue Kaspersky Safe Kids.",
			smsLangHY: "Վաղուց չենք հանդիպել: Լիցքավորեք հաշիվը՝ շարունակելու Kaspersky Safe Kids:",
			smsLangRU: "Давно не встречались. Пополните баланс для продолжения Kaspersky Safe Kids.",
		},
		scOff: {
			smsLangEN: "{{.ProductLabel}} service deactivated.",
			smsLangHY: "{{.ProductLabel}} ծառայությունն ապաակտիվացված է:",
			smsLangRU: "Услуга {{.ProductLabel}} деактивирована.",
		},
	}
}

func mergeSmsTemplates(cfg map[string]map[string]string) map[string]map[string]string {
	out := defaultSmsTemplates()
	for scenario, locales := range cfg {
		if out[scenario] == nil {
			out[scenario] = make(map[string]string)
		}
		for loc, text := range locales {
			if strings.TrimSpace(text) != "" {
				out[scenario][localeOrDefault(loc)] = text
			}
		}
	}
	return out
}

func (s *Viva) initSMS() {
	tplText := s.smpp.Template
	if tplText == "" {
		tplText = "{{.ProductName}}{{ if .Quantity }} for {{ pluralizeEn .Quantity \"device\" \"devices\" }}{{ end }}\n" +
			"Activation code: {{.ActivationCode}}\n" +
			"Download link: {{.DownloadURL}}"
	}
	s.activationTpl, _ = template.New("activation").Funcs(tplext.Funcs).Parse(tplText)
	s.smppSender = NewSmppSender(s.smpp)

	raw := mergeSmsTemplates(s.smpp.SmsTemplates)
	s.scenarioTpl = make(map[string]map[string]*template.Template)
	for scenario, locales := range raw {
		s.scenarioTpl[scenario] = make(map[string]*template.Template)
		for loc, text := range locales {
			tpl, err := template.New(scenario + "_" + loc).Parse(text)
			if err != nil {
				logger.Warn().Str("scenario", scenario).Str("locale", loc).Err(err).Msg("skip invalid sms template")
				continue
			}
			s.scenarioTpl[scenario][loc] = tpl
		}
	}
}

func (s *Viva) sendScenario(phone, scenario, locale string, data SmsTplData) error {
	if s.smppSender == nil {
		return fmt.Errorf("smppSender is nil")
	}
	lang := localeOrDefault(locale)
	if strings.TrimSpace(data.ProductLabel) == "" {
		data.ProductLabel = "Kaspersky Safe Kids"
	}
	if strings.TrimSpace(data.ProductName) == "" {
		data.ProductName = data.ProductLabel
	}

	byLoc := s.scenarioTpl[scenario]
	if byLoc == nil {
		return fmt.Errorf("unknown sms scenario %q", scenario)
	}
	tpl := byLoc[lang]
	if tpl == nil {
		tpl = byLoc[smsLangHY]
	}
	if tpl == nil {
		return fmt.Errorf("no template for scenario %q locale %q", scenario, lang)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}
	body := strings.TrimSpace(buf.String())
	if body == "" {
		return fmt.Errorf("empty sms body for scenario %s", scenario)
	}

	logger.Info().Str("sender", "SMPP").Str("scenario", scenario).Str("msisdn", phone).Msg("SMS")
	return s.smppSender.Send(phone, body)
}
