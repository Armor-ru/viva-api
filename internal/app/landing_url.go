package viva_api

import (
	"fmt"
	"net/url"
	"strings"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

func (s *viva) buildLandingConfirmURL(phone, productName, productCode, lang string) (string, error) {
	base := strings.TrimSpace(s.landingConfirmURL)
	if base == "" {
		return "", fmt.Errorf("landingConfirmURL is not configured")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("landingConfirmURL: %w", err)
	}
	q := u.Query()
	q.Set("phone", normalizePhone(phone))
	q.Set("productName", strings.TrimSpace(productName))
	q.Set("productCode", strings.TrimSpace(productCode))
	q.Set("lang", strings.TrimSpace(lang))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *viva) sendOtpLandingSMS(phone string, product catalog.Product, lang string) error {
	productCode := product.GetExternalID()
	catalogProduct, err := s.productForExternalID(FlowActivation, "7", productCode)
	if err != nil {
		return err
	}
	landingURL, err := s.buildLandingConfirmURL(phone, vivaProductName(productCode), productCode, lang)
	if err != nil {
		flowError(FlowActivation, "7").Err(err).Str("phone", phone).Msg("landing URL")
		return err
	}
	data := map[string]interface{}{
		"Phone":       phone,
		"ExternalID":  productCode,
		"ShortNumber": product.GetShortNumber(),
		"Language":    lang,
		"LandingURL":  landingURL,
	}
	text := catalogProduct.GetNotify(NotifyOtpLanding, data, lang)
	if text == "" {
		return fmt.Errorf("%s notification template is empty", NotifyOtpLanding)
	}
	flowInfo(FlowActivation, "7").Str("phone", phone).Str("lang", lang).Msg("sending OTP landing link SMS")
	if err := s.notify(phone, text); err != nil {
		return fmt.Errorf("send %s sms: %w", NotifyOtpLanding, err)
	}
	return nil
}
