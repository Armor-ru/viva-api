package viva_api

import (
	"fmt"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

// MO на 1020 с текстом, отличным от 1, STOP, RUS, ARM, ENG → СМС №13 (справка).

func unknownStep1ReceiveOn1020(mo inboundMO) error {
	if mo.Phone == "" {
		return fmt.Errorf("subscriber phone is empty")
	}
	if !mo.isUnknownCommandOn1020() {
		return fmt.Errorf("expected unknown command on %s, got %q on %q", activationShortNumber, mo.Text, mo.ShortNumber)
	}
	flowInfo(FlowUnknown, "1").
		Str("phone", mo.Phone).
		Str("shortNumber", mo.ShortNumber).
		Str("text", mo.Text).
		Msg("inbound unknown command on 1020 via SMPP")
	return nil
}

func (s *viva) handleUnknownOn1020(mo inboundMO) {
	if err := unknownStep1ReceiveOn1020(mo); err != nil {
		flowError(FlowUnknown, "1").Err(err).
			Str("phone", mo.Phone).
			Str("shortNumber", mo.ShortNumber).
			Str("text", mo.Text).
			Msg("invalid unknown-command MO")
		return
	}

	product, err := s.productForShortNumber(FlowUnknown, "2", mo.ShortNumber)
	if err != nil {
		return
	}

	if err := s.sendUnknownCommandSMS(mo, product); err != nil {
		flowError(FlowUnknown, "3").Err(err).Str("phone", mo.Phone).Msg("failed to send SMS #13")
	}
}

func (s *viva) sendUnknownCommandSMS(mo inboundMO, product catalog.Product) error {
	phone := normalizePhone(mo.Phone)
	lang := s.resolveLang(phone, product.GetDefaultLanguage())
	flowInfo(FlowUnknown, "3").Str("phone", phone).Str("lang", lang).Msg("sending SMS #13 via SMPP")
	return s.notifyFromProduct(phone, product, NotifyUnknownCommand, product.GetDefaultLanguage(), nil)
}

