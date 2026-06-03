package viva_api

import (
	"strings"
	"testing"
)

func TestScenario8_Step1_ReceivesUnknownOn1020(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"HELP", "2", "hello", "  x  "} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: text}
			if !mo.isUnknownCommandOn1020() {
				t.Fatal("expected unknown command MO")
			}
			if err := unknownStep1ReceiveOn1020(mo); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScenario8_Step1_NotUnknownForKnownCommands(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"1", "STOP", "RUS", "ARM", "ENG", "stop", "rus"} {
		mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: text}
		if mo.isUnknownCommandOn1020() {
			t.Fatalf("%q must not be scenario 8", text)
		}
		if err := unknownStep1ReceiveOn1020(mo); err == nil {
			t.Fatalf("%q: expected error", text)
		}
	}
}

func TestScenario8_Step1_NotUnknownOnOtherShortNumber(t *testing.T) {
	t.Parallel()

	mo := inboundMO{Phone: "37477600552", ShortNumber: "9999", Text: "HELP"}
	if mo.isUnknownCommandOn1020() {
		t.Fatal("unknown scenario is only for 1020")
	}
}

func TestScenario8_Step2_ProductByShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.productForShortNumber(FlowUnknown, "2", "1020")
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" || product.GetShortNumber() != "1020" {
		t.Fatalf("product = %+v", product)
	}
}

func TestScenario8_Step2_UnknownShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	if _, err := v.productForShortNumber(FlowUnknown, "2", "9999"); err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario8_Step2_HandleLoadsProduct(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	v.handleUnknownOn1020(inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "HELP"})
}

func TestScenario8_Step3_SendsSMS13_Russian(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{ussdTransport: notifyTr, catalog: loadTestCatalog(t), langStore: NewLangStore("ru")}
	product, _ := v.productForShortNumber(FlowUnknown, "2", "1020")
	mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "HELP"}

	if err := v.sendUnknownCommandSMS(mo, product); err != nil {
		t.Fatal(err)
	}
	want := "Уважаемый абонент, для подключения услуги"
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if len(text) < len(want) || text[:len(want)] != want {
		t.Fatalf("sms #13 = %q", text)
	}
}

func TestScenario8_Step3_SendsSMS13_English(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	store := NewLangStore("ru")
	store.Set("37477600552", "en")
	v := &viva{ussdTransport: notifyTr, catalog: loadTestCatalog(t), langStore: store}
	product, _ := v.productForShortNumber(FlowUnknown, "2", "1020")

	if err := v.sendUnknownCommandSMS(
		inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "HELP"},
		product,
	); err != nil {
		t.Fatal(err)
	}
	want := "Dear subscriber, to activate the service"
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if !strings.HasPrefix(text, want) {
		t.Fatalf("sms #13 = %q", text)
	}
}

func TestScenario8_FullFlow_HELP_SendsSMS13(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
	}

	v.handleUnknownOn1020(inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "HELP"})

	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected 1 SMS, got %d", len(notifyTr.sendCalls))
	}
	if !strings.Contains(notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string), "ARM, RUS или ENG") {
		t.Fatal("expected Russian help SMS #13")
	}
}

func TestScenario8_Step1_UssdHandlerRoutesUnknown(t *testing.T) {
	t.Parallel()

	intTr := &fakeTransport{}
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:    intTr,
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
	}

	v.ussdHandler(&handlerCtx{payload: inboundSMSDTO{
		Phone: "37477600552", ShortNumber: "1020", Text: "HELP",
	}})

	if len(intTr.sendCalls) != 0 {
		t.Fatalf("scenario 8 must not publish to NATS, got %d", len(intTr.sendCalls))
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected SMS #13, got %d", len(notifyTr.sendCalls))
	}
}

func TestScenario8_Step1_HandleUnknown(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	v.handleUnknownOn1020(inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "HELP"})
}

func TestInboundMO_UnknownDoesNotMatchKnown(t *testing.T) {
	t.Parallel()

	mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "HELP"}
	if !mo.isUnknownCommandOn1020() {
		t.Fatal("expected unknown")
	}
	if mo.isKnownCommandOn1020() {
		t.Fatal("HELP must not be a known command")
	}
}
