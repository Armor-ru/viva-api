package viva_api

import (
	"fmt"
	"net/url"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

type UssdRequest struct {
	Phone       string `json:"sourceAddr" validate:"required"`
	ShortNumber string `json:"destinationAddr" validate:"required"`
	Text        string `json:"shortMessage" validate:"required"`
}

func (s *Viva) buildLandingConfirmURL(phone, productName, productCode, lang string) (string, error) {
	base := strings.TrimSpace(s.landingConfirmURL)
	if base == "" {
		return "", fmt.Errorf("landing confirm url is empty")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("landing confirm url: %w", err)
	}
	q := u.Query()
	q.Set("phone", strings.TrimPrefix(strings.TrimSpace(phone), "+"))
	q.Set("productName", strings.TrimSpace(productName))
	q.Set("productCode", strings.TrimSpace(productCode))
	if lang = strings.TrimSpace(lang); lang != "" {
		q.Set("lang", lang)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *Viva) ussdHandler(ctx types.HandlerContext) {
	var req UssdRequest
	ctx.Data(&req)
	req.Phone = strings.TrimSpace(strings.TrimPrefix(req.Phone, "+"))
	req.ShortNumber = strings.TrimSpace(req.ShortNumber)
	req.Text = strings.TrimSpace(strings.ToUpper(req.Text))
	logger.Info().
		Str("phone", req.Phone).
		Str("shortNumber", req.ShortNumber).
		Str("text", req.Text).
		Msg("smpp mo received")
	s.ussd(req)
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

			landingURL, err := s.buildLandingConfirmURL(req.Phone, product.ExternalId, product.ExternalId, lang)
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
