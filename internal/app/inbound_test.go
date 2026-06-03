package viva_api

import "testing"

func TestParseInboundMO_ActivationOneOn1020(t *testing.T) {
	t.Parallel()

	mo := parseInboundMO(inboundSMSDTO{
		Phone:       "+37477600552",
		ShortNumber: "1020",
		Text:        " 1 ",
	})
	if mo.Phone != "37477600552" {
		t.Fatalf("phone = %q", mo.Phone)
	}
	if mo.ShortNumber != "1020" {
		t.Fatalf("shortNumber = %q", mo.ShortNumber)
	}
	if mo.Text != "1" {
		t.Fatalf("text = %q", mo.Text)
	}
	if !mo.isActivationOne() {
		t.Fatal("expected activation trigger for step 1")
	}
}

func TestParseInboundMO_StopOn1020(t *testing.T) {
	t.Parallel()

	mo := parseInboundMO(inboundSMSDTO{
		Phone:       "37477600552",
		ShortNumber: "1020",
		Text:        " stop ",
	})
	if mo.Text != "stop" {
		t.Fatalf("text = %q", mo.Text)
	}
	if !mo.isStopOn1020() {
		t.Fatal("expected STOP on 1020")
	}
	if mo.isActivationOne() {
		t.Fatal("STOP must not match activation")
	}
}

func TestParseInboundMO_LanguageChangeOn1020(t *testing.T) {
	t.Parallel()

	mo := parseInboundMO(inboundSMSDTO{
		Phone:       "37477600552",
		ShortNumber: "1020",
		Text:        " rus ",
	})
	if mo.Text != "rus" {
		t.Fatalf("text = %q", mo.Text)
	}
	if !mo.isLanguageChangeOn1020() {
		t.Fatal("expected RUS/ARM/ENG on 1020")
	}
	if mo.languageCommandCode() != "ru" {
		t.Fatalf("lang = %q", mo.languageCommandCode())
	}
}

func TestParseInboundMO_UnknownCommandOn1020(t *testing.T) {
	t.Parallel()

	mo := parseInboundMO(inboundSMSDTO{
		Phone:       "37477600552",
		ShortNumber: "1020",
		Text:        " help ",
	})
	if mo.Text != "help" {
		t.Fatalf("text = %q", mo.Text)
	}
	if !mo.isUnknownCommandOn1020() {
		t.Fatal("expected unknown command on 1020")
	}
}

func TestParseInboundMO_NormalizesShortNumber(t *testing.T) {
	t.Parallel()

	mo := parseInboundMO(inboundSMSDTO{
		Phone:       "37477600552",
		ShortNumber: "+1020",
		Text:        "1",
	})
	if !mo.isActivationOne() {
		t.Fatalf("shortNumber = %q", mo.ShortNumber)
	}
}
