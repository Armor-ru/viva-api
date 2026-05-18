package viva_api

import (
	"bytes"
	"fmt"
	"text/template"
)

type smsTemplates struct {
	templates map[string]*template.Template
}

func newSMSTemplates(templates map[string]map[string]string) (*smsTemplates, error) {
	if len(templates) == 0 {
		return nil, fmt.Errorf("sms templates not configured")
	}

	st := &smsTemplates{
		templates: make(map[string]*template.Template),
	}
	for scenario, langs := range templates {
		for lang, text := range langs {
			key := scenario + ":" + lang
			tmpl, err := template.New(key).Parse(text)
			if err != nil {
				return nil, fmt.Errorf("parse template %s: %w", key, err)
			}
			st.templates[key] = tmpl
		}
	}
	return st, nil
}

func mustSMSTemplates(templates map[string]map[string]string) *smsTemplates {
	st, err := newSMSTemplates(templates)
	if err != nil {
		panic("sms templates: " + err.Error())
	}
	return st
}

func (st *smsTemplates) render(scenario, lang string, data SmsData) string {
	lang = localeOrDefault(lang)

	for _, l := range []string{lang, "en"} {
		tmpl, ok := st.templates[scenario+":"+l]
		if !ok {
			continue
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			continue
		}
		return buf.String()
	}
	return ""
}
