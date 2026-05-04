package viva_api

import (
	"fmt"
	"strings"
)

// Поддерживаемые коды для SMS (SMPP) и префиксов /landing/{код}/...
const (
	SmsLangEN = "en"
	SmsLangRU = "ru"
	SmsLangHY = "hy"
)

// LocaleOrDefault нормализует язык для SMS; пустое значение → ru (как было до i18n).
func LocaleOrDefault(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "en", "eng", "english":
		return SmsLangEN
	case "ru", "rus", "russian":
		return SmsLangRU
	case "hy", "arm", "hye", "armenian":
		return SmsLangHY
	default:
		return SmsLangRU
	}
}

// ParseLocalePath проверяет сегмент пути :locale (строго en|ru|hy).
func ParseLocalePath(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", fmt.Errorf("locale required")
	}
	switch s {
	case SmsLangEN, "eng", "english":
		return SmsLangEN, nil
	case SmsLangRU, "rus", "russian":
		return SmsLangRU, nil
	case SmsLangHY, "arm", "hye", "armenian":
		return SmsLangHY, nil
	default:
		return "", fmt.Errorf("unknown locale %q: use en, ru, hy", s)
	}
}

func landingPickLocale(pathLocale, bodyLocale string) (string, error) {
	if strings.TrimSpace(pathLocale) != "" {
		return ParseLocalePath(pathLocale)
	}
	return LocaleOrDefault(bodyLocale), nil
}
