package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func (s *Viva) notify(phone, from, text, notifyKey string) error {

	if s.ussdTransport == nil {
		return errs.WrapWithFields(
			fmt.Errorf("ussdTransport is not configured"),
			nil,
		)
	}
	phone = strings.TrimSpace(strings.TrimPrefix(phone, "+"))
	from = strings.TrimSpace(from)
	text = strings.TrimSpace(text)
	if phone == "" || from == "" || text == "" {
		return errs.WrapWithFields(
			fmt.Errorf("phone, from or text is empty"),
			map[string]interface{}{"phone": phone, "from": from, "text": text},
		)
	}
	_, err := s.ussdTransport.Send("", map[string]interface{}{
		"from": from,
		"to":   phone,
		"text": text,
	}, types.SendOptions{})
	if err != nil {
		return errs.WrapWithFields(
			fmt.Errorf("ussd transport send failed, %w", err),
			map[string]interface{}{"phone": phone, "from": from, "text": text},
		)
	}
	log := logger.Info().
		Str("phone", phone).
		Str("from", from).
		Int("textLen", len(text)).
		Str("text", text)
	if notifyKey != "" {
		log = log.Str("notify", notifyKey)
	}
	log.Msg("smpp mt sent")
	return nil

}

func (s *Viva) sendServiceDeactivated(phone, productCode string) error {
	product, err := s.catalog.GetProductByExternalId(productCode)
	if err != nil {
		return err
	}
	lang := s.GetLang(phone)
	if lang == "" {
		lang = "ru"
	}
	return s.sendProductNotify(product, phone, "service_deactivated", map[string]interface{}{
		"Phone":       phone,
		"ShortNumber": product.ShortNumber,
		"ExternalID":  product.ExternalId,
	}, lang)
}

func (s *Viva) sendProductNotify(product *Product, phone, key string, data map[string]interface{}, lang string) error {
	text := product.GetNotify(key, data, lang)
	if text == "" {
		return errs.WrapWithFields(
			fmt.Errorf("notification template is empty"),
			map[string]interface{}{
				"shortNumber": product.ShortNumber,
				"key":         key,
				"lang":        lang,
			},
		)
	}
	return s.notify(phone, product.ShortNumber, text, key)
}
