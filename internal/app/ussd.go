package viva_api

import (
	"fmt"
	"net/url"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
)

func (s *Viva) buildLandingConfirmURL(base, phone string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		return "", fmt.Errorf("landing confirm url is empty")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("landing confirm url %w", err)
	}
	q := u.Query()
	q.Set("phone", strings.TrimPrefix(strings.TrimSpace(phone), "+"))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *Viva) ussd(req UssdRequest) {
	product, err := s.catalog.GetProductByShortNumber(req.ShortNumber)
	if err != nil {
		logAppError(err, "get product by short number failed")
		return
	}
	lang := s.GetLang(req.Phone)
	if lang == "" {
		lang = "ru"
	}
	switch req.Text {
	case "1":
		orderID := s.getOrderId(req.Phone, product.ExternalId)
		order, err := s.getOrder(orderID)
		exists := err == nil && strings.TrimSpace(order.ID) != ""

		data := map[string]interface{}{
			"Phone":       req.Phone,
			"ShortNumber": req.ShortNumber,
			"ExternalID":  product.ExternalId,
		}

		var key string

		switch {
		case exists && s.isActiveOrder(order):
			key = "already_active"

		case !exists:
			if _, err := s.landingInit(SubscriptionInitRequest{
				PhoneNum:    req.Phone,
				ProductName: product.ExternalId,
				Lang:        lang,
			}); err != nil {
				logAppError(err, "init subscription failed")
				return
			}

			landingURL, err := s.buildLandingConfirmURL(product.LandingConfirmURL, req.Phone)
			if err != nil {
				logAppError(err, "build landing url failed")
				return
			}
			data["LandingURL"] = landingURL
			key = "otp_landing"

		default:
			logAppError(
				errs.WrapWithFields(
					fmt.Errorf("order exists but not active"),
					map[string]interface{}{
						"orderId":     orderID,
						"phone":       req.Phone,
						"shortNumber": req.ShortNumber,
						"externalId":  product.ExternalId,
					},
				),
				"skip init subscription",
			)
			return
		}
		if err := s.sendProductNotify(product, req.Phone, key, data, lang); err != nil {
			logAppError(err, "notify failed")
		}
	case "STOP":
		productCode := product.ExternalId
		orderID := s.getOrderId(req.Phone, productCode)
		order, err := s.getOrder(orderID)
		exists := err == nil && strings.TrimSpace(order.ID) != ""

		data := map[string]interface{}{
			"Phone":       req.Phone,
			"ShortNumber": req.ShortNumber,
			"ExternalID":  productCode,
		}
		key := "service_deactivated"
		if exists && !s.isActiveOrder(order) {
			key = "already_deactivated"
		} else if err := s.removeSubscription(req.Phone, productCode); err != nil {
			logAppError(err, "remove subscription failed")
			return
		}
		if err := s.sendProductNotify(product, req.Phone, key, data, lang); err != nil {
			logAppError(err, "notify failed")
		}
	case "RUS", "ENG", "ARM":
		switch req.Text {
		case "RUS":
			lang = "ru"
		case "ENG":
			lang = "en"
		case "ARM":
			lang = "arm"
		}
		s.SetLang(req.Phone, lang)
		if err := s.sendProductNotify(product, req.Phone, "language_changed", map[string]interface{}{
			"Phone": req.Phone, "ShortNumber": req.ShortNumber, "Language": lang,
		}, lang); err != nil {
			logAppError(err, "notify failed")
		}
	default:
		if err := s.sendProductNotify(product, req.Phone, "unknown_command", map[string]interface{}{
			"Phone": req.Phone, "ShortNumber": req.ShortNumber, "Text": req.Text,
		}, lang); err != nil {
			logAppError(err, "notify failed")
		}
	}
}
