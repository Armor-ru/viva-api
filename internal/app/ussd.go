package viva_api

import (
	"fmt"
	"net/url"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"github.com/spf13/cast"
)

func (s *Viva) buildLandingConfirmURL(base, phone, lang string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		return "", fmt.Errorf("landing confirm url is empty")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("landing confirm url failed, %w", err)
	}
	q := u.Query()
	q.Set("phone", strings.TrimPrefix(strings.TrimSpace(phone), "+"))
	lang = normalizeLangCode(lang)
	if lang == "" {
		lang = defaultLang
	}
	q.Set("lang", lang)
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
		lang = defaultLang
	}
	switch req.Text {
	case "1":
		data := map[string]interface{}{
			"Phone":       req.Phone,
			"ShortNumber": req.ShortNumber,
			"ExternalID":  product.ExternalId,
		}

		if _, err := s.landingInit(SubscriptionInitRequest{
			PhoneNum:    req.Phone,
			ProductName: product.ExternalId,
			Lang:        lang,
		}); err != nil {
			if fields := errs.Fields(err); fields != nil {
				if cast.ToInt(fields["resultCode"]) == 7 {
					logUssdInitDone(req.Phone, product.ExternalId, 7, "already_active")
					if err := s.sendProductNotify(product, req.Phone, "already_active", data, lang); err != nil {
						logAppError(err, "notify failed")
					}
					return
				}
			}
			logAppError(err, "init subscription failed")
			return
		}

		landingURL, err := s.buildLandingConfirmURL(product.LandingConfirmURL, req.Phone, lang)
		if err != nil {
			logAppError(err, "build landing url failed")
			return
		}
		data["LandingURL"] = landingURL
		if err := s.sendProductNotify(product, req.Phone, "otp_landing", data, lang); err != nil {
			logAppError(err, "notify failed")
			return
		}
		logUssdInitDone(req.Phone, product.ExternalId, 0, "landing_sms")
	case "STOP":
		data := map[string]interface{}{
			"Phone":       req.Phone,
			"ShortNumber": req.ShortNumber,
			"ExternalID":  product.ExternalId,
		}
		if err := s.removeSubscription(req.Phone, product.ExternalId); err != nil {
			if fields := errs.Fields(err); fields != nil && cast.ToInt(fields["resultCode"]) == 20 {
				logUssdStopDone(req.Phone, product.ExternalId, "already_deactivated")
				if err := s.sendProductNotify(product, req.Phone, "already_deactivated", data, lang); err != nil {
					logAppError(err, "notify failed")
				}
				return
			}
			logAppError(err, "remove subscription failed")
			return
		}
		if err := s.sendProductNotify(product, req.Phone, "service_deactivated", data, lang); err != nil {
			logAppError(err, "notify failed")
			return
		}
		logUssdStopDone(req.Phone, product.ExternalId, "service_deactivated")
	case "RUS", "ENG", "ARM":
		switch req.Text {
		case "RUS":
			lang = "ru"
		case "ENG":
			lang = "en"
		case "ARM":
			lang = "hy"
		}
		s.SetLang(req.Phone, lang)
		if err := s.sendProductNotify(product, req.Phone, "language_changed", map[string]interface{}{
			"Phone": req.Phone, "ShortNumber": req.ShortNumber, "Language": lang,
		}, lang); err != nil {
			logAppError(err, "notify failed")
			return
		}
		logUssdLangDone(req.Phone, product.ExternalId, lang)
	default:
		if err := s.sendProductNotify(product, req.Phone, "unknown_command", map[string]interface{}{
			"Phone": req.Phone, "ShortNumber": req.ShortNumber, "Text": req.Text,
		}, lang); err != nil {
			logAppError(err, "notify failed")
			return
		}
		logUssdUnknownDone(req.Phone, product.ExternalId, req.Text, lang)
	}
}
