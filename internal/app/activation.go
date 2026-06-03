package viva_api

import (
	"strings"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

// MO «1» на 1020: новая подписка (сценарий 1) или уже активна (сценарий 2, СМС №8).
// Цепочка: каталог → order/get → Viva InitSubscription → СМС со ссылкой на лендинг (OTP на лендинге).
// Подтверждение на лендинге → order/create → СМС №2 и №3 по order/completed.

func (s *viva) handleActivationOne(mo inboundMO) {
	activationStep1(mo)

	product, err := s.activationStep2(mo)
	if err != nil {
		return
	}

	if mo.Phone == "" {
		flowError(FlowActivation, "3").Msg("subscriber phone is empty")
		return
	}
	if strings.TrimSpace(s.accountId) == "" {
		flowError(FlowActivation, "3").Msg("accountId is not configured")
		return
	}

	productCode := product.GetExternalID()
	orderID := s.getOrderId(mo.Phone, productCode)
	flowInfo(FlowActivation, "3").
		Str("orderId", orderID).
		Str("phone", mo.Phone).
		Str("productCode", productCode).
		Msg("orderId formed from product and phone")

	if s.intTransport == nil {
		flowError(FlowActivation, "4").Msg("intTransport is not configured")
		return
	}

	lookup := s.activationStep4Lookup(orderID)
	if activationOrderAlreadyActive(lookup) {
		if err := s.sendAlreadyActiveSMS(mo, product); err != nil {
			flowError(FlowActivation, "5").Err(err).Str("orderId", orderID).Msg("already-active branch failed")
		}
		return
	}
	if lookup.Exists {
		flowInfo(FlowActivation, "5").
			Str("orderId", lookup.Order.ID).
			Str("status", lookup.Order.Status).
			Bool("active", lookup.Active).
			Msg("order/get — order exists but is not active")
		return
	}

	flowInfo(FlowActivation, "5").Str("orderId", orderID).Msg("order/get — order does not exist")

	productName := vivaProductName(productCode)
	if err := s.vivaInitSubscription(mo.Phone, productName); err != nil {
		return
	}

	lang := s.resolveLang(mo.Phone, product.GetDefaultLanguage())
	if err := s.sendOtpLandingSMS(mo.Phone, product, lang); err != nil {
		flowError(FlowActivation, "7").Err(err).Str("orderId", orderID).Msg("OTP landing SMS failed")
	}
}

func activationStep1(mo inboundMO) {
	flowInfo(FlowActivation, "1").
		Str("phone", mo.Phone).
		Str("shortNumber", mo.ShortNumber).
		Str("text", mo.Text).
		Msg("inbound SMS \"1\" on 1020 via SMPP")
}

func (s *viva) activationStep2(mo inboundMO) (catalog.Product, error) {
	return s.productForShortNumber(FlowActivation, "2", mo.ShortNumber)
}

func (s *viva) activationStep4Lookup(orderID string) orderLookup {
	flowInfo(FlowActivation, "4").Str("orderId", orderID).Msg("publishing order/get")
	return s.lookupOrder(orderID)
}

func activationOrderAlreadyActive(lookup orderLookup) bool {
	if !lookup.Exists || !lookup.Active {
		return false
	}
	flowInfo(FlowActivation, "5").
		Str("orderId", lookup.Order.ID).
		Str("status", lookup.Order.Status).
		Msg("order/get — order exists and is active")
	return true
}

func (s *viva) sendAlreadyActiveSMS(mo inboundMO, product catalog.Product) error {
	return s.sendAlreadyActiveSMSForPhone(mo.Phone, product)
}

func (s *viva) sendAlreadyActiveSMSForPhone(phone string, product catalog.Product) error {
	phone = normalizePhone(phone)
	lang := s.resolveLang(phone, product.GetDefaultLanguage())
	flowInfo(FlowActivation, "7").Str("phone", phone).Str("lang", lang).Msg("sending SMS #8 via SMPP")
	return s.notifyFromProduct(phone, product, NotifyAlreadyActive, product.GetDefaultLanguage(), nil)
}
