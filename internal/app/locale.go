package viva_api

import (
	"fmt"
	"strings"
)

const (
	smsLangEN = "en"
	smsLangRU = "ru"
	smsLangHY = "hy"
)

func localeOrDefault(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "en", "eng", "english":
		return smsLangEN
	case "ru", "rus", "russian":
		return smsLangRU
	case "hy", "arm", "hye", "armenian":
		return smsLangHY
	default:
		return smsLangHY
	}
}

func parseLocalePath(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", fmt.Errorf("locale required")
	}
	switch s {
	case smsLangEN, "eng", "english":
		return smsLangEN, nil
	case smsLangRU, "rus", "russian":
		return smsLangRU, nil
	case smsLangHY, "arm", "hye", "armenian":
		return smsLangHY, nil
	default:
		return "", fmt.Errorf("unknown locale %q: use en, ru, hy", s)
	}
}

func landingPickLocale(pathLocale, bodyLocale string) (string, error) {
	if strings.TrimSpace(pathLocale) != "" {
		return parseLocalePath(pathLocale)
	}
	return localeOrDefault(bodyLocale), nil
}
