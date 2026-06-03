package viva_api

import (
	"strings"
	"testing"
)

func TestSendActivationSMS_Step13_RussianWelcomeAndLicense(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	store := loadTestCatalog(t)
	product, err := store.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}

	v := &viva{ussdTransport: notifyTr}
	data := map[string]interface{}{
		"Link":           "https://download.example/safekids",
		"ActivationCode": "KEY-99",
	}
	if err := v.sendActivationSMS("37477600552", "ru", product, data); err != nil {
		t.Fatal(err)
	}
	if len(notifyTr.sendCalls) != 2 {
		t.Fatalf("sms count = %d", len(notifyTr.sendCalls))
	}

	welcome := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if !strings.Contains(welcome, "Поздравляем! Услуга Kaspersky Safe Kids подключена") {
		t.Fatalf("welcome = %q", welcome)
	}
	if !strings.Contains(welcome, "https://download.example/safekids") {
		t.Fatalf("welcome missing link: %q", welcome)
	}

	license := notifyTr.sendCalls[1].msg.(map[string]interface{})["text"].(string)
	if license != "Kaspersky Safe Kids: KEY-99" {
		t.Fatalf("license = %q", license)
	}
}

func TestSendActivationSMS_Step13_English(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	store := loadTestCatalog(t)
	product, _ := store.GetProductByExternalId("SAFEKID")
	v := &viva{ussdTransport: notifyTr}

	err := v.sendActivationSMS("37477600552", "en", product, map[string]interface{}{
		"Link": "https://x.test", "ActivationCode": "K1",
	})
	if err != nil {
		t.Fatal(err)
	}
	welcome := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if !strings.Contains(welcome, "Congratulations! Kaspersky Safe Kids is activated") {
		t.Fatalf("welcome = %q", welcome)
	}
}
