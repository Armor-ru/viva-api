package viva_api

import (
	"fmt"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

// notifyFromProduct рендерит шаблон из каталога и отправляет СМС по SMPP.
// extra объединяется с данными шаблона (Phone, ShortNumber, ExternalID, Language).
// Если передан explicitLang — используется он; иначе resolveLang(phone, fallbackLang).
func (s *viva) notifyFromProduct(phone string, product catalog.Product, notifyKey, fallbackLang string, extra map[string]interface{}, explicitLang ...string) error {
	phone = normalizePhone(phone)
	if phone == "" {
		return fmt.Errorf("phone is empty")
	}

	var lang string
	if len(explicitLang) > 0 && explicitLang[0] != "" {
		lang = normalizeLang(explicitLang[0])
	} else {
		lang = s.resolveLang(phone, fallbackLang)
	}
	if lang == "" {
		return fmt.Errorf("language is empty")
	}

	data := map[string]interface{}{
		"Phone":       phone,
		"ShortNumber": product.GetShortNumber(),
		"ExternalID":  product.GetExternalID(),
		"Language":    lang,
	}
	for k, v := range extra {
		data[k] = v
	}

	text := product.GetNotify(notifyKey, data, lang)
	if text == "" {
		return fmt.Errorf("%s notification template is empty for lang %q", notifyKey, lang)
	}

	if err := s.notify(phone, text); err != nil {
		return fmt.Errorf("send %s sms: %w", notifyKey, err)
	}
	return nil
}
