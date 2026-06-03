package catalog

import (
	"bytes"
	"strings"
	"text/template"
)

func renderTemplate(text string, data map[string]interface{}) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}
	tpl, err := template.New("notify").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
