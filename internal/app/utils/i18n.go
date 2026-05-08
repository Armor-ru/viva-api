package utils

import (
	"fmt"
	"strings"
)

const (
	SmsLangEN = "en"
	SmsLangRU = "ru"
	SmsLangHY = "hy"
)

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
		return SmsLangHY
	}
}

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

func LandingPickLocale(pathLocale, bodyLocale string) (string, error) {
	if strings.TrimSpace(pathLocale) != "" {
		return ParseLocalePath(pathLocale)
	}
	return LocaleOrDefault(bodyLocale), nil
}
