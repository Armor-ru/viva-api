package viva_api

import (
	"strings"
	"testing"
)

func TestScenario7_Step1_ReceivesRusArmEngOn1020(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"RUS", "ARM", "ENG", "rus", "eng", "arm"} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: text}
			if !mo.isLanguageChangeOn1020() {
				t.Fatal("expected language change MO")
			}
			if err := langChangeStep1ReceiveOn1020(mo); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScenario7_Step1_LanguageCommandCode(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"RUS": "ru",
		"ENG": "en",
		"ARM": "arm",
	}
	for text, want := range cases {
		got := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: text}.languageCommandCode()
		if got != want {
			t.Fatalf("%s: got %q want %q", text, got, want)
		}
	}
}

func TestScenario7_Step1_RejectsOtherTexts(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"1", "STOP", "RU", "ENGLISH"} {
		mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: text}
		if mo.isLanguageChangeOn1020() {
			t.Fatalf("%q must not match scenario 7", text)
		}
		if err := langChangeStep1ReceiveOn1020(mo); err == nil {
			t.Fatalf("%q: expected error", text)
		}
	}
}

func TestScenario7_Step1_RejectsWrongShortNumber(t *testing.T) {
	t.Parallel()

	mo := inboundMO{Phone: "37477600552", ShortNumber: "9999", Text: "RUS"}
	if err := langChangeStep1ReceiveOn1020(mo); err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario7_Step2_ProductByShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.productForShortNumber(FlowLang, "2", "1020")
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" || product.GetShortNumber() != "1020" {
		t.Fatalf("product = %+v", product)
	}
}

func TestScenario7_Step2_UnknownShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	if _, err := v.productForShortNumber(FlowLang, "2", "9999"); err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario7_Step2_HandleLoadsProduct(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t), langStore: NewLangStore("ru")}
	v.handleLanguageOn1020(inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "ENG"})
}

func TestScenario7_Step3_SavesLanguagePreference(t *testing.T) {
	t.Parallel()

	store := NewLangStore("ru")
	v := &viva{langStore: store}
	mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "ENG"}

	lang, err := v.saveLanguagePreference(mo)
	if err != nil || lang != "en" {
		t.Fatalf("lang=%q err=%v", lang, err)
	}
	if store.Get("37477600552") != "en" {
		t.Fatal("preference not stored")
	}
}

func TestScenario7_Step3_UsedByScenario5SMS(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	store := NewLangStore("ru")
	v := &viva{
		ussdTransport: &fakeTransport{},
		catalog:         loadTestCatalog(t),
		langStore:       store,
	}
	product, _ := v.productForShortNumber(FlowLang, "2", "1020")
	if _, err := v.saveLanguagePreference(inboundMO{Phone: phone, ShortNumber: "1020", Text: "ENG"}); err != nil {
		t.Fatal(err)
	}

	if err := v.sendStopDeactivatedSMS(
		inboundMO{Phone: phone, ShortNumber: "1020", Text: "STOP"},
		product,
	); err != nil {
		t.Fatal(err)
	}
	want := "The service is deactivated"
	text := v.ussdTransport.(*fakeTransport).sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if len(text) < len(want) || text[:len(want)] != want {
		t.Fatalf("sms #6 = %q", text)
	}
}

func TestScenario7_Step4_SendsSMS10_Russian(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{ussdTransport: notifyTr, catalog: loadTestCatalog(t), langStore: NewLangStore("ru")}
	product, _ := v.productForShortNumber(FlowLang, "2", "1020")
	mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "RUS"}

	if err := v.sendLanguageChangedSMS(mo, product, "ru"); err != nil {
		t.Fatal(err)
	}
	want := "Уважаемый абонент, язык SMS-сообщений изменен на русский"
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if len(text) < len(want) || text[:len(want)] != want {
		t.Fatalf("sms #10 = %q", text)
	}
}

func TestScenario7_Step4_SendsSMS11_Armenian(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{ussdTransport: notifyTr, catalog: loadTestCatalog(t), langStore: NewLangStore("ru")}
	product, _ := v.productForShortNumber(FlowLang, "2", "1020")
	mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "ARM"}

	if err := v.sendLanguageChangedSMS(mo, product, "arm"); err != nil {
		t.Fatal(err)
	}
	want := "Հարգելի բաժանորդ, SMS հաղորդագրությունների լեզուն փոխվել է հայերենի"
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if len(text) < len(want) || text[:len(want)] != want {
		t.Fatalf("sms #11 = %q", text)
	}
}

func TestScenario7_Step4_SendsSMS12_English(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{ussdTransport: notifyTr, catalog: loadTestCatalog(t), langStore: NewLangStore("ru")}
	product, _ := v.productForShortNumber(FlowLang, "2", "1020")
	mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "ENG"}

	if err := v.sendLanguageChangedSMS(mo, product, "en"); err != nil {
		t.Fatal(err)
	}
	want := "Dear subscriber, the language of SMS messages has been changed to English"
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if len(text) < len(want) || text[:len(want)] != want {
		t.Fatalf("sms #12 = %q", text)
	}
}

func TestScenario7_FullFlow_RUS_SendsSMS10(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
	}

	v.handleLanguageOn1020(inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "RUS"})

	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected 1 SMS, got %d", len(notifyTr.sendCalls))
	}
	want := "изменен на русский"
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if !strings.Contains(text, want) {
		t.Fatalf("sms = %q", text)
	}
	if v.langStore.Get("37477600552") != "ru" {
		t.Fatalf("stored lang = %q", v.langStore.Get("37477600552"))
	}
}

func TestScenario7_Step1_UssdHandlerRoutesLanguageChange(t *testing.T) {
	t.Parallel()

	vivaAPI := newVivaAPITestDouble(t)
	intTr := &fakeTransport{}
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:    intTr,
		vivaClient:      vivaAPI.Client(),
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
		accountId:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	v.ussdHandler(&handlerCtx{payload: inboundSMSDTO{
		Phone: "37477600552", ShortNumber: "1020", Text: "RUS",
	}})

	if vivaAPI.RemoveCalls != 0 {
		t.Fatalf("scenario 7 must not call Viva, got %d", vivaAPI.RemoveCalls)
	}
	if len(intTr.sendCalls) != 0 {
		t.Fatalf("scenario 7 must not publish to NATS, got %d", len(intTr.sendCalls))
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected SMS #10, got %d", len(notifyTr.sendCalls))
	}
}

func TestInboundMO_LanguageChangeDoesNotMatchActivationOrStop(t *testing.T) {
	t.Parallel()

	mo := inboundMO{Phone: "37477600552", ShortNumber: "1020", Text: "ENG"}
	if !mo.isLanguageChangeOn1020() {
		t.Fatal("expected language change")
	}
	if mo.isActivationOne() || mo.isStopOn1020() {
		t.Fatal("ENG must not match activation or STOP")
	}
}
