package catalog

import (
	"fmt"
	"strings"
	"text/template"

	"git.dev.armlab.pro/armor/sds-go/pkg/tplext"
)

func renderTemplate(tpl string, data map[string]interface{}) (string, error) {
	compiled, err := template.New("catalog-notify").Funcs(tplext.Funcs).Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf strings.Builder
	if err := compiled.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}
