package viva_api

import (
	"fmt"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

// MO RUS / ARM / ENG на 1020: сохранение языка (TTL) и отправка СМС №10–№12.

func langChangeStep1ReceiveOn1020(mo inboundMO) error {
	if mo.Phone == "" {
		return fmt.Errorf("subscriber phone is empty")
	}
	if !mo.isLanguageChangeOn1020() {
		return fmt.Errorf("expected RUS, ARM, or ENG on %s, got %q on %q", activationShortNumber, mo.Text, mo.ShortNumber)
	}
	flowInfo(FlowLang, "1").
		Str("phone", mo.Phone).
		Str("shortNumber", mo.ShortNumber).
		Str("text", mo.Text).
		Str("lang", mo.languageCommandCode()).
		Msg("inbound SMS RUS/ARM/ENG on 1020 via SMPP")
	return nil
}

func (s *viva) handleLanguageOn1020(mo inboundMO) {
	if err := langChangeStep1ReceiveOn1020(mo); err != nil {
		flowError(FlowLang, "1").Err(err).
			Str("phone", mo.Phone).
			Str("shortNumber", mo.ShortNumber).
			Str("text", mo.Text).
			Msg("invalid language-change MO")
		return
	}

	product, err := s.productForShortNumber(FlowLang, "2", mo.ShortNumber)
	if err != nil {
		return
	}

	lang, err := s.saveLanguagePreference(mo)
	if err != nil {
		flowError(FlowLang, "3").Err(err).Str("phone", mo.Phone).Msg("failed to save language preference")
		return
	}

	if err := s.sendLanguageChangedSMS(mo, product, lang); err != nil {
		flowError(FlowLang, "4").Err(err).Str("phone", mo.Phone).Str("lang", lang).Msg("failed to send confirmation SMS")
	}
}

func (s *viva) saveLanguagePreference(mo inboundMO) (string, error) {
	lang := mo.languageCommandCode()
	if lang == "" {
		return "", fmt.Errorf("language command code is empty")
	}
	if s.langStore == nil {
		return "", fmt.Errorf("lang store is not configured")
	}
	phone := normalizePhone(mo.Phone)
	s.langStore.Set(phone, lang)

	flowInfo(FlowLang, "3").Str("phone", phone).Str("lang", lang).Msg("language preference saved (TTL)")

	return lang, nil
}

func langConfirmationSMSNumber(lang string) string {
	switch normalizeLang(lang) {
	case "ru":
		return "10"
	case "arm":
		return "11"
	case "en":
		return "12"
	default:
		return ""
	}
}

func (s *viva) sendLanguageChangedSMS(mo inboundMO, product catalog.Product, lang string) error {
	phone := normalizePhone(mo.Phone)
	lang = normalizeLang(lang)
	smsNo := langConfirmationSMSNumber(lang)
	flowInfo(FlowLang, "4").
		Str("phone", phone).
		Str("lang", lang).
		Str("sms", smsNo).
		Msg("sending language-changed SMS via SMPP")
	return s.notifyFromProduct(phone, product, NotifyLanguageChanged, "", nil, lang)
}

