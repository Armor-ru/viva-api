package viva_api

import (
	"strings"
	"testing"
	"time"
)

func TestCatalog_LoadAndGetNotify(t *testing.T) {
	t.Parallel()

	c := NewCatalog()
	if err := c.Load("../../catalog"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	p, err := c.GetProductByShortNumber("1020")
	if err != nil {
		t.Fatalf("GetProductByShortNumber() error = %v", err)
	}
	if p.ExternalId != "SAFEKID" {
		t.Fatalf("ExternalId = %q, want SAFEKID", p.ExternalId)
	}
	if p.LandingConfirmURL != "https://safekids.viva.am/landing_1/" {
		t.Fatalf("LandingConfirmURL = %q, want https://safekids.viva.am/landing_1/", p.LandingConfirmURL)
	}
	if got := p.GetNotify("otp_landing", map[string]interface{}{"LandingURL": "https://safekids.viva.am/landing_1/?phone=37493215362"}, "ru"); got != "Вы пытаетесь зарегистрироваться в сервисе Kaspersky Safe Kids. Для регистрации перейдите по ссылке https://safekids.viva.am/landing_1/?phone=37493215362 и введите только что полученный от Viva SMS-код." {
		t.Fatalf("GetNotify(otp_landing) = %q", got)
	}
	if got := p.GetNotify("language_changed", nil, "ru"); got != "Уважаемый абонент, язык SMS-сообщений изменен на русский. Инфо: support@kaspersky.com" {
		t.Fatalf("GetNotify(language_changed) = %q", got)
	}
	if got := p.GetNotify("otp_landing", map[string]interface{}{"LandingURL": "https://safekids.viva.am/landing_1/?phone=37493215362"}); !strings.Contains(got, "Դուք փորձում եք") {
		t.Fatalf("GetNotify(otp_landing, default) = %q", got)
	}

	p2, err := c.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatalf("GetProductByExternalId() error = %v", err)
	}
	if got := p2.GetNotify("license", map[string]interface{}{"ActivationCode": "ABC123"}, "ru"); got != "“Kaspersky Safe Kids”: ABC123" {
		t.Fatalf("GetNotify(license) = %q", got)
	}
	end := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	got := p.GetNotify("trial_expires", map[string]interface{}{"EndTime": &end}, "ru")
	if got == "" || !strings.Contains(got, "2026-06-05") {
		t.Fatalf("GetNotify(trial_expires) = %q", got)
	}
}

func TestGetNotify_pluralize(t *testing.T) {
	t.Parallel()

	p := &Product{
		Notifications: map[string]interface{}{
			"ru": map[string]interface{}{
				"devices": "{{ pluralize .Quantity \"устройство\" \"устройства\" \"устройств\" }}",
			},
		},
	}
	if got := p.GetNotify("devices", map[string]interface{}{"Quantity": 2}, "ru"); got != "2 устройства" {
		t.Fatalf("GetNotify(pluralize) = %q", got)
	}
}
