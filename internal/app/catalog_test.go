package viva_api

import "testing"

func TestCatalog_LoadAndGetNotify(t *testing.T) {
	t.Parallel()

	c := NewCatalog()
	if err := c.Load("../../catalog"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := c.SetDefaultLang("ru"); err != nil {
		t.Fatalf("SetDefaultLang() error = %v", err)
	}

	p, err := c.GetProductByShortNumber("1024")
	if err != nil {
		t.Fatalf("GetProductByShortNumber() error = %v", err)
	}
	if got := p.GetNotify("new", map[string]interface{}{"ExternalID": "SAFEKID"}, "ru"); got != "Подключено SAFEKID" {
		t.Fatalf("GetNotify(new) = %q", got)
	}

	p2, err := c.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatalf("GetProductByExtermalId() error = %v", err)
	}
	if got := p2.GetNotify("cancel", map[string]interface{}{"ExternalID": "SAFEKID"}, "ru"); got != "Отключено SAFEKID" {
		t.Fatalf("GetNotify(cancel) = %q", got)
	}
}
