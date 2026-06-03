package viva_api

import (
	"testing"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestScenario6_Step1_ReceivesStopOn1020(t *testing.T) {
	t.Parallel()

	if err := stopStep1ReceiveOn1020(inboundMO{
		Phone: "37477600552", ShortNumber: "1020", Text: "STOP",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestScenario6_Step1_RejectsNonStop(t *testing.T) {
	t.Parallel()

	if err := stopStep1ReceiveOn1020(inboundMO{
		Phone: "37477600552", ShortNumber: "1020", Text: "1",
	}); err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario6_Step2_ProductByShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.productForShortNumber(FlowStop, "2", "1020")
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" || product.GetShortNumber() != "1020" {
		t.Fatalf("product = %+v", product)
	}
}

func TestScenario6_Step5_OrderAlreadyDeactivated(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-time.Hour)
	lookup := orderLookup{
		Exists: true,
		Active: false,
		Order:  Order{ID: "ord-1", Status: string(types.OrderStatusListCompleted), EndTime: &past},
	}
	if !stopOrderAlreadyDeactivated(lookup) {
		t.Fatal("expected already deactivated")
	}
}

func TestScenario6_Step5_ActiveOrder_NotScenario6(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(time.Hour)
	lookup := orderLookup{
		Exists: true,
		Active: true,
		Order:  Order{ID: "ord-1", Status: string(types.OrderStatusListCompleted), EndTime: &future},
	}
	if stopOrderAlreadyDeactivated(lookup) {
		t.Fatal("active order should use scenario 5")
	}
}

func TestScenario6_Step5_RoutesToScenario6_NotVivaRemove(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	orderID := (&viva{accountId: accountID}).getOrderId(phone, "SAFEKID")
	past := time.Now().Add(-time.Hour)

	vivaAPI := newVivaAPITestDouble(t)
	intTr := &orderGetTransport{
		order: Order{
			ID: orderID, Status: string(types.OrderStatusListCompleted), EndTime: &past,
			Fields: map[string]interface{}{"phone": phone},
		},
	}
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:    intTr,
		vivaClient:      vivaAPI.Client(),
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
		accountId:       accountID,
	}

	v.handleStopOn1020(inboundMO{Phone: phone, ShortNumber: "1020", Text: "STOP"})

	if vivaAPI.RemoveCalls != 0 {
		t.Fatalf("scenario 6 should not call RemoveSubscription, got %d", vivaAPI.RemoveCalls)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected SMS #9, got %d", len(notifyTr.sendCalls))
	}
	wantSub := "Уважаемый абонент, услуга уже отключена"
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if len(text) < len(wantSub) || text[:len(wantSub)] != wantSub {
		t.Fatalf("sms #9 = %q", text)
	}
}

func TestScenario6_Step6_SendsSMS9_Russian(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
	}
	product, _ := v.productForShortNumber(FlowStop, "2", "1020")

	if err := v.sendStopAlreadyOffSMS(
		inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "STOP"},
		product,
	); err != nil {
		t.Fatal(err)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("sms count = %d", len(notifyTr.sendCalls))
	}
}

func TestScenario6_Step4_PublishesOrderGet(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	orderID := (&viva{accountId: accountID}).getOrderId(phone, "SAFEKID")

	intTr := &orderGetTransport{
		order: Order{
			ID: orderID, Status: "completed",
			Fields: map[string]interface{}{"phone": phone},
		},
	}
	v := &viva{intTransport: intTr, accountId: accountID}

	lookup, err := v.stopOrderGetLookup(orderID)
	if err != nil {
		t.Fatal(err)
	}
	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("calls = %+v", intTr.sendCalls)
	}
	if !lookup.Exists {
		t.Fatal("expected order to exist")
	}
}

func TestScenario6_Step3_OrderIdFromProductAndPhone(t *testing.T) {
	t.Parallel()

	v := &viva{
		accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		catalog:   loadTestCatalog(t),
	}
	product, err := v.productForShortNumber(FlowStop, "2", "1020")
	if err != nil {
		t.Fatal(err)
	}

	orderID, err := v.stopFormOrderID("37477600552", product)
	if err != nil {
		t.Fatal(err)
	}
	want := v.getOrderId("37477600552", "SAFEKID")
	if orderID != want || orderID == "" {
		t.Fatalf("orderId = %q want %q", orderID, want)
	}
}

func TestScenario6_Step2_UnknownShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	_, err := v.productForShortNumber(FlowStop, "2", "9999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario6_Step1_IsStopOn1020(t *testing.T) {
	t.Parallel()

	mo := parseInboundMO(inboundSMSDTO{Phone: "37477600552", ShortNumber: "1020", Text: "STOP"})
	if !mo.isStopOn1020() {
		t.Fatal("expected STOP on 1020")
	}
}
