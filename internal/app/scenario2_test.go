package viva_api

import (
	"testing"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestScenario2_Step2_ProductByShortNumber1020(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.activationStep2(inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" || product.GetShortNumber() != "1020" {
		t.Fatalf("product = %s / %s", product.GetExternalID(), product.GetShortNumber())
	}
}

func TestScenario2_Step2_UnknownShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	_, err := v.activationStep2(inboundMO{ShortNumber: "9999", Text: "1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario2_ActiveOrder_SendsSMS8_NoVivaNoCreate(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	v := &viva{accountId: accountID}
	orderID := v.getOrderId(phone, "SAFEKID")
	future := time.Now().Add(time.Hour)

	vivaAPI := newVivaAPITestDouble(t)
	intTr := &orderGetTransport{
		order: Order{ID: orderID, Status: string(types.OrderStatusListCompleted), EndTime: &future},
	}
	notifyTr := &fakeTransport{}
	v = &viva{
		intTransport:    intTr,
		vivaClient:      vivaAPI.Client(),
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		accountId:       accountID,
		langStore:       NewLangStore("ru"),
	}

	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})

	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("nats = %+v", intTr.sendCalls)
	}
	if vivaAPI.InitCalls != 0 || vivaAPI.ConfirmCalls != 0 {
		t.Fatal("viva must not be called in scenario 2")
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected 1 SMS #8, got %d", len(notifyTr.sendCalls))
	}
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	want := "Уважаемый абонент, услуга уже подключена. Инфо:"
	if text != want {
		t.Fatalf("sms #8 = %q, want %q", text, want)
	}
}

func TestScenario2_Step7_SMS8_English(t *testing.T) {
	t.Parallel()

	store := loadTestCatalog(t)
	product, _ := store.GetProductByExternalId("SAFEKID")
	notifyTr := &fakeTransport{}
	v := &viva{ussdTransport: notifyTr, langStore: NewLangStore("ru")}
	v.langStore.Set("37477600552", "en")

	err := v.sendAlreadyActiveSMSForPhone("37477600552", product)
	if err != nil {
		t.Fatal(err)
	}
	want := "Dear subscriber, the service is already activated. Info:"
	if got := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string); got != want {
		t.Fatalf("sms #8 = %q", got)
	}
}

func TestScenario2_OrderExistsNotActive_NoSMS8(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	v := &viva{accountId: accountID}
	orderID := v.getOrderId(phone, "SAFEKID")

	intTr := &orderGetTransport{order: Order{ID: orderID, Status: "pending"}}
	notifyTr := &fakeTransport{}
	v = &viva{
		intTransport:    intTr,
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		accountId:       accountID,
	}

	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})

	if len(notifyTr.sendCalls) != 0 {
		t.Fatalf("expected no SMS, got %d", len(notifyTr.sendCalls))
	}
}

func TestScenario2_Step1_InboundMO_TriggersActivation(t *testing.T) {
	t.Parallel()

	mo := parseInboundMO(inboundSMSDTO{
		Phone: "37477600552", ShortNumber: "1020", Text: "1",
	})
	if !mo.isActivationOne() {
		t.Fatal("expected activation entry for scenario 1/2 step 1")
	}
}

func TestScenario2_Step1_UssdHandler_DoesNotRunGenericUSSD(t *testing.T) {
	t.Parallel()

	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	notifyTr := &fakeTransport{}
	vivaAPI := newVivaAPITestDouble(t)
	v := &viva{
		catalog:         loadTestCatalog(t),
		accountId:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		intTransport:    intTr,
		vivaClient:      vivaAPI.Client(),
		ussdTransport: notifyTr,
		langStore:       NewLangStore("ru"),
	}

	v.ussdHandler(&handlerCtx{payload: inboundSMSDTO{
		Phone: "37477600552", ShortNumber: "1020", Text: "1",
	}})

	if len(intTr.sendCalls) == 0 {
		t.Fatal("expected activation flow to continue past step 1")
	}
	if intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("first NATS topic = %q, want order/get", intTr.sendCalls[0].topic)
	}
}
