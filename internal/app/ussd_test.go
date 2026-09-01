package viva_api

import (
	"net/url"
	"testing"
)

func TestBuildLandingConfirmURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		phone    string
		lang     string
		wantLang string
	}{
		{name: "russian", base: "https://safekids.viva.am/", phone: "+37493215362", lang: "ru", wantLang: "ru"},
		{name: "english", base: "https://safekids.viva.am/", phone: "37493215362", lang: "en", wantLang: "en"},
		{name: "armenian", base: "https://safekids.viva.am/", phone: "37493215362", lang: "hy", wantLang: "hy"},
		{name: "default", base: "https://safekids.viva.am/", phone: "37493215362", lang: "", wantLang: "hy"},
	}

	v := &Viva{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			landingURL, err := v.buildLandingConfirmURL(tt.base, tt.phone, tt.lang)
			if err != nil {
				t.Fatalf("buildLandingConfirmURL() error = %v", err)
			}

			parsed, err := url.Parse(landingURL)
			if err != nil {
				t.Fatalf("url.Parse() error = %v", err)
			}
			if got := parsed.Query().Get("phone"); got != "37493215362" {
				t.Fatalf("phone = %q, want 37493215362", got)
			}
			if got := parsed.Query().Get("lang"); got != tt.wantLang {
				t.Fatalf("lang = %q, want %q", got, tt.wantLang)
			}
		})
	}
}
