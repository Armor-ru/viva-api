package viva

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/Armor-ru/sds-go/pkg/tplext"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
)

func TestSmsTemplatePluralize(t *testing.T) {
	tplText :=
		"{{.ProductName}}{{ if .Quantity }} for {{ pluralizeEn .Quantity \"device\" \"devices\" }}{{ end }}\n" +
			"Activation code: {{.ActivationCode}}\n" +
			"Download link: {{.DownloadURL}}"

	tpl, err := template.New("sms").Funcs(tplext.Funcs).Parse(tplText)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var b bytes.Buffer
	err = tpl.Execute(&b, viva_api.SmsData{
		ProductName:    "SafeKids",
		Quantity:       22,
		ActivationCode: "ABC",
		DownloadURL:    "http://x",
	})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}

	got := b.String()
	if wantSub := "SafeKids for 22 devices"; !contains(got, wantSub) {
		t.Fatalf("got %q; want contains %q", got, wantSub)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()))
}
