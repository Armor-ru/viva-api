package viva_api

import (
	"strings"
	"time"
)

type LangStore map[string]langEntry

type langEntry struct {
	lang      string
	expiresAt time.Time
}

const (
	langPreferenceTTL = 24 * time.Hour
	defaultLang       = "hy"
)

func normalizeLangCode(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "hy", "en", "ru":
		return lang
	default:
		return ""
	}
}

func (s *Viva) GetLang(phone string) string {
	phone = strings.TrimSpace(strings.TrimPrefix(phone, "+"))
	if phone == "" {
		return ""
	}
	e, ok := s.langStore[phone]
	if !ok || time.Now().After(e.expiresAt) {
		return ""
	}
	return normalizeLangCode(e.lang)
}

func (s *Viva) SetLang(phone, lang string) {
	phone = strings.TrimSpace(strings.TrimPrefix(phone, "+"))
	lang = normalizeLangCode(lang)
	if phone == "" || lang == "" {
		return
	}
	s.langStore[phone] = langEntry{
		lang:      lang,
		expiresAt: time.Now().Add(langPreferenceTTL),
	}
}
