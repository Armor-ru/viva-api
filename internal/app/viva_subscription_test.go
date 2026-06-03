package viva_api

import (
	"strings"
	"testing"
)

func TestHandleActivationOne_Step6_VivaInitAndLandingSMS(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	vivaAPI := newVivaAPITestDouble(t)
	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:      intTr,
		ussdTransport:     notifyTr,
		vivaClient:        vivaAPI.Client(),
		catalog:           loadTestCatalog(t),
		accountId:         accountID,
		landingConfirmURL: "https://landing.test/confirm",
	}

	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})

	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("nats calls = %+v", intTr.sendCalls)
	}
	if vivaAPI.InitCalls != 1 || vivaAPI.ConfirmCalls != 0 {
		t.Fatalf("viva init=%d confirm=%d", vivaAPI.InitCalls, vivaAPI.ConfirmCalls)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected 1 landing SMS, got %d", len(notifyTr.sendCalls))
	}
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "https://landing.test/confirm") {
		t.Fatalf("sms = %q", text)
	}
}

func TestHandleActivationOne_Step6_SkippedWhenOrderExistsNotActive(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	v := &viva{accountId: accountID}
	orderID := v.getOrderId(phone, "SAFEKID")
	vivaAPI := newVivaAPITestDouble(t)
	intTr := &orderGetTransport{order: Order{ID: orderID, Status: "pending"}}
	v = &viva{
		intTransport: intTr,
		vivaClient:   vivaAPI.Client(),
		catalog:      loadTestCatalog(t),
		accountId:    accountID,
	}

	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})

	if vivaAPI.InitCalls != 0 || vivaAPI.ConfirmCalls != 0 {
		t.Fatalf("viva should not be called when order exists but not active")
	}
}
