package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

// MO STOP на 1020: отключение (сценарий 5, СМС №6) или уже отключено (сценарий 6, СМС №9).

func stopStep1ReceiveOn1020(mo inboundMO) error {
	if mo.Phone == "" {
		return fmt.Errorf("subscriber phone is empty")
	}
	if !mo.isStopOn1020() {
		return fmt.Errorf("expected STOP on %s, got %q on %q", activationShortNumber, mo.Text, mo.ShortNumber)
	}
	flowInfo(FlowStop, "1").
		Str("phone", mo.Phone).
		Str("shortNumber", mo.ShortNumber).
		Str("text", mo.Text).
		Msg("inbound SMS STOP on 1020 via SMPP")
	return nil
}

func (s *viva) handleStopOn1020(mo inboundMO) {
	if err := stopStep1ReceiveOn1020(mo); err != nil {
		flowError(FlowStop, "1").Err(err).
			Str("phone", mo.Phone).
			Str("shortNumber", mo.ShortNumber).
			Str("text", mo.Text).
			Msg("invalid STOP MO")
		return
	}

	product, err := s.productForShortNumber(FlowStop, "2", mo.ShortNumber)
	if err != nil {
		return
	}

	orderID, err := s.stopFormOrderID(mo.Phone, product)
	if err != nil {
		return
	}

	lookup, err := s.stopOrderGetLookup(orderID)
	if err != nil {
		flowError(FlowStop, "4").Err(err).Str("orderId", orderID).Msg("order/get failed")
		return
	}

	if stopOrderAlreadyDeactivated(lookup) {
		if err := s.runStopAlreadyOff(mo, product); err != nil {
			flowError(FlowStop, "6").Err(err).Str("phone", mo.Phone).Msg("already-off branch failed")
		}
		return
	}

	if lookup.Exists && lookup.Active {
		if err := s.runStopDeactivate(mo, product); err != nil {
			flowError(FlowStop, "4").Err(err).Str("phone", mo.Phone).Msg("deactivate branch failed")
		}
		return
	}

	flowInfo(FlowStop, "5").Str("orderId", orderID).Msg("order/get — no active order, running deactivate")
	if err := s.runStopDeactivate(mo, product); err != nil {
		flowError(FlowStop, "5").Err(err).Str("phone", mo.Phone).Msg("deactivate branch failed")
	}
}

func (s *viva) stopFormOrderID(phone string, product catalog.Product) (string, error) {
	phone = normalizePhone(phone)
	if phone == "" {
		return "", fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(s.accountId) == "" {
		return "", fmt.Errorf("accountId is not configured")
	}
	orderID := s.getOrderId(phone, product.GetExternalID())
	flowInfo(FlowStop, "3").
		Str("orderId", orderID).
		Str("phone", phone).
		Str("externalId", product.GetExternalID()).
		Msg("orderId formed")
	return orderID, nil
}

func (s *viva) stopOrderGetLookup(orderID string) (orderLookup, error) {
	if s.intTransport == nil {
		return orderLookup{}, fmt.Errorf("intTransport is not configured")
	}
	flowInfo(FlowStop, "4").Str("orderId", orderID).Msg("publishing order/get")
	return s.lookupOrder(orderID), nil
}

func stopOrderAlreadyDeactivated(lookup orderLookup) bool {
	if !lookup.Exists || lookup.Active {
		return false
	}
	flowInfo(FlowStop, "5").
		Str("orderId", lookup.Order.ID).
		Str("status", lookup.Order.Status).
		Msg("order/get — already deactivated")
	return true
}

func (s *viva) runStopDeactivate(mo inboundMO, product catalog.Product) error {
	if err := s.removeVivaSubscription(mo.Phone, product.GetExternalID()); err != nil {
		return err
	}
	return s.sendStopDeactivatedSMS(mo, product)
}

func (s *viva) runStopAlreadyOff(mo inboundMO, product catalog.Product) error {
	return s.sendStopAlreadyOffSMS(mo, product)
}

func (s *viva) sendStopDeactivatedSMS(mo inboundMO, product catalog.Product) error {
	phone := normalizePhone(mo.Phone)
	lang := s.resolveLang(phone, product.GetDefaultLanguage())
	flowInfo(FlowStop, "4").Str("phone", phone).Str("lang", lang).Msg("sending SMS #6 via SMPP")

	if err := s.notifyFromProduct(phone, product, NotifyServiceDeactivated, product.GetDefaultLanguage(), nil); err != nil {
		return err
	}

	s.paidWelcomeStore().Clear(phone, product.GetExternalID())
	return nil
}

func (s *viva) sendStopAlreadyOffSMS(mo inboundMO, product catalog.Product) error {
	phone := normalizePhone(mo.Phone)
	lang := s.resolveLang(phone, product.GetDefaultLanguage())
	flowInfo(FlowStop, "6").Str("phone", phone).Str("lang", lang).Msg("sending SMS #9 via SMPP")
	return s.notifyFromProduct(phone, product, NotifyAlreadyDeactivated, product.GetDefaultLanguage(), nil)
}
