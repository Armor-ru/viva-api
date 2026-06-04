package viva_api

import (
	"strings"
	"testing"
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
	if got := p.GetNotify("language_changed", nil, "ru"); got != "Уважаемый абонент, язык SMS-сообщений изменен на русский. Инфо:" {
		t.Fatalf("GetNotify(language_changed) = %q", got)
	}

	p2, err := c.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatalf("GetProductByExternalId() error = %v", err)
	}
	if got := p2.GetNotify("license", map[string]interface{}{"ActivationCode": "ABC123"}, "ru"); got != "Kaspersky Safe Kids: ABC123" {
		t.Fatalf("GetNotify(license) = %q", got)
	}
	got := p.GetNotify("trial_expires", map[string]interface{}{"ExpiresAt": "05.06.2026 12:00"}, "ru")
	if got == "" || !strings.Contains(got, "05.06.2026 12:00") {
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
