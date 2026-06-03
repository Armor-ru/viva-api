package viva_api

import (
	"encoding/json"
	"testing"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestInitHandlers_SubscribesOrderExpiresOnNATS(t *testing.T) {
	t.Parallel()

	intTr := &subscribeRecordingTransport{}
	dir := t.TempDir()
	writeTestCatalog(t, dir)
	_ = New(
		WithIntTransport(intTr),
		WithCatalogDir(dir),
	)

	found := false
	for _, topic := range intTr.topics {
		if topic == orderExpiresSubscribePath {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("subscriptions = %v", intTr.topics)
	}
}

func TestParseOrderExpiresJSON_OrderExpireRequest(t *testing.T) {
	t.Parallel()

	raw, _ := json.Marshal(types.OrderExpireRequest{
		ID: "ord-1", Status: "completed", EndTime: "2026-06-10T00:00:00Z",
	})
	p, err := ParseOrderExpiresJSON(raw)
	if err != nil || p.ID != "ord-1" || p.EndTime == "" {
		t.Fatalf("payload = %+v err = %v", p, err)
	}
}

func TestParseOrderExpiresJSON_OrderResponse(t *testing.T) {
	t.Parallel()

	end := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	raw, _ := json.Marshal(types.OrderResponse{
		ID: "ord-2", Status: "completed", EndTime: &end,
		Fields: map[string]interface{}{"phone": "37477600552"},
	})
	p, err := ParseOrderExpiresJSON(raw)
	if err != nil || p.ID != "ord-2" {
		t.Fatalf("payload = %+v err = %v", p, err)
	}
}

func TestOrderExpiresHandler_Step2_ReceivesOrderExpireRequest(t *testing.T) {
	t.Parallel()

	v := &viva{}
	v.orderExpiresHandler(&orderExpiresCtx{
		payload: types.OrderExpireRequest{
			ID: "order-1", Status: "completed", EndTime: "2026-06-10T12:00:00Z",
		},
	})
}

func TestOrderExpiresHandler_Step2_ReceivesFullOrderResponse(t *testing.T) {
	t.Parallel()

	end := time.Now().Add(24 * time.Hour)
	v := &viva{}
	v.orderExpiresHandler(&orderExpiresCtx{
		order: types.OrderResponse{
			ID: "order-2", Status: "completed", EndTime: &end,
			CustomData: map[string]interface{}{"productCode": "SAFEKID"},
		},
	})
}

func TestScenario3_Step3_ProductByExternalId_FromPayload(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.productForExternalID(FlowTrial, "3", "SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	if product.GetShortNumber() != "1020" {
		t.Fatalf("shortNumber = %q", product.GetShortNumber())
	}
}

func TestScenario3_Step3_ProductByExternalId_ViaOrderGet(t *testing.T) {
	t.Parallel()

	orderID := "order-expire-1"
	end := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	intTr := &orderGetTransport{
		order: Order{
			ID: orderID, Status: "completed", EndTime: &end,
			Fields:     map[string]interface{}{"phone": "37477600552"},
			CustomData: map[string]interface{}{"productCode": "SAFEKID", "lang": "ru"},
		},
	}
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:    intTr,
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
	}

	err := v.handleOrderExpires(orderExpiresPayload{ID: orderID})
	if err != nil {
		t.Fatal(err)
	}
	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("calls = %+v", intTr.sendCalls)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected SMS #5, got %d sends", len(notifyTr.sendCalls))
	}
}

func TestScenario3_Step4_SendsSMS5_Russian(t *testing.T) {
	t.Parallel()

	end := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	intTr := &orderGetTransport{
		order: Order{
			ID: "order-expire-1", Status: "completed", EndTime: &end,
			Fields:     map[string]interface{}{"phone": "37477600552"},
			CustomData: map[string]interface{}{"productCode": "SAFEKID", "lang": "ru"},
		},
	}
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:    intTr,
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
	}

	v.orderExpiresHandler(&orderExpiresCtx{
		order: types.OrderResponse{
			ID: "order-expire-1", Status: "completed", EndTime: &end,
			Fields:     map[string]interface{}{"phone": "37477600552"},
			CustomData: map[string]interface{}{"productCode": "SAFEKID", "lang": "ru"},
		},
	})

	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("sms count = %d", len(notifyTr.sendCalls))
	}
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if notifyTr.sendCalls[0].msg.(map[string]interface{})["to"].(string) != "37477600552" {
		t.Fatal("wrong recipient")
	}
	wantSub := "Уважаемый абонент, 10.06.2026 12:00 пробный период истекает"
	if len(text) < len(wantSub) || text[:len(wantSub)] != wantSub {
		t.Fatalf("sms #5 = %q", text)
	}
}

func TestScenario3_Step3_UnknownExternalId(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	_, err := v.productForExternalID(FlowTrial, "3", "UNKNOWN")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleOrderExpires_EmptyID(t *testing.T) {
	t.Parallel()

	v := &viva{}
	err := v.handleOrderExpires(orderExpiresPayload{})
	if err == nil {
		t.Fatal("expected error")
	}
}

type orderExpiresCtx struct {
	payload types.OrderExpireRequest
	order   types.OrderResponse
	useOrder bool
}

func (c *orderExpiresCtx) Data(dest interface{}) {
	switch d := dest.(type) {
	case *types.OrderExpireRequest:
		if c.useOrder {
			*d = types.OrderExpireRequest{}
			return
		}
		*d = c.payload
	case *types.OrderResponse:
		*d = c.order
	}
}

func (c *orderExpiresCtx) Param(string) string          { return "" }
func (c *orderExpiresCtx) Headers(interface{})        {}
func (c *orderExpiresCtx) Response(interface{}) error { return nil }
func (c *orderExpiresCtx) Error(error) error          { return nil }
