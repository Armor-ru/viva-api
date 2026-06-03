package viva_api

import (
	"testing"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

func activeOrderGetTransport(phone string) *orderGetTransport {
	v := &viva{accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}
	orderID := v.getOrderId(phone, "SAFEKID")
	future := time.Now().Add(time.Hour)
	return &orderGetTransport{
		order: Order{
			ID: orderID, Status: string(types.OrderStatusListCompleted), EndTime: &future,
			Fields: map[string]interface{}{"phone": phone},
		},
	}
}

func TestScenario5_Step1_ReceivesStopOn1020(t *testing.T) {
	t.Parallel()

	err := stopStep1ReceiveOn1020(inboundMO{
		Phone: "37477600552", ShortNumber: "1020", Text: "STOP",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestScenario5_Step1_RejectsWrongShortNumber(t *testing.T) {
	t.Parallel()

	err := stopStep1ReceiveOn1020(inboundMO{
		Phone: "37477600552", ShortNumber: "9999", Text: "STOP",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario5_Step3_ProductByShortNumber(t *testing.T) {
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

func TestScenario5_Step3_UnknownShortNumber(t *testing.T) {
	t.Parallel()

	vivaAPI := newVivaAPITestDouble(t)
	v := &viva{vivaClient: vivaAPI.Client(), catalog: loadTestCatalog(t)}
	v.handleStopOn1020(inboundMO{Phone: "37477600552", ShortNumber: "9999", Text: "STOP"})

	if vivaAPI.RemoveCalls != 0 {
		t.Fatalf("expected no RemoveSubscription, got %d", vivaAPI.RemoveCalls)
	}
}

func TestScenario5_Step4_SendsSMS6_Russian(t *testing.T) {
	t.Parallel()

	vivaAPI := newVivaAPITestDouble(t)
	notifyTr := &fakeTransport{}
	v := &viva{
		vivaClient:      vivaAPI.Client(),
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
		paidWelcomeSent: NewPaidWelcomeStore(),
	}

	if err := v.sendStopDeactivatedSMS(
		inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "STOP"},
		mustLoadProduct(t),
	); err != nil {
		t.Fatal(err)
	}

	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	wantSub := "Услуга деактивирована. Для повторной подписки"
	if len(text) < len(wantSub) || text[:len(wantSub)] != wantSub {
		t.Fatalf("sms #6 = %q", text)
	}
}

func mustLoadProduct(t *testing.T) catalog.Product {
	t.Helper()
	v := &viva{catalog: loadTestCatalog(t)}
	p, err := v.productForShortNumber(FlowStop, "2", "1020")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestScenario5_Step2_CallsRemoveSubscription(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	vivaAPI := newVivaAPITestDouble(t)
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:    activeOrderGetTransport(phone),
		vivaClient:      vivaAPI.Client(),
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
		accountId:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	v.handleStopOn1020(inboundMO{Phone: phone, ShortNumber: "1020", Text: "STOP"})

	if vivaAPI.RemoveCalls != 1 {
		t.Fatalf("remove calls = %d", vivaAPI.RemoveCalls)
	}
	if vivaAPI.LastRemove.Phone != "37477600552" || vivaAPI.LastRemove.Product != "SAFEKID" {
		t.Fatalf("remove args = %+v", vivaAPI.LastRemove)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected SMS #6, got %d", len(notifyTr.sendCalls))
	}
}

func TestScenario5_Step2_SkipsWhenVivaNotConfigured(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	v.handleStopOn1020(inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "STOP"})
}

func TestScenario5_Step1_UssdHandlerRoutesStop(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	vivaAPI := newVivaAPITestDouble(t)
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:    activeOrderGetTransport(phone),
		vivaClient:      vivaAPI.Client(),
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		accountId:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	v.ussdHandler(&handlerCtx{payload: inboundSMSDTO{
		Phone: smppField(phone), ShortNumber: "1020", Text: "STOP",
	}})

	if vivaAPI.RemoveCalls != 1 {
		t.Fatalf("expected RemoveSubscription, got %d calls", vivaAPI.RemoveCalls)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected SMS #6, got %d", len(notifyTr.sendCalls))
	}
}
