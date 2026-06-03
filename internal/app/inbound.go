package viva_api

import "strings"

func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+")
	return s
}

func normalizeShortNumber(s string) string {
	return strings.TrimSpace(strings.TrimPrefix(s, "+"))
}

const activationShortNumber = "1020"

// inboundMO — нормализованное MO-сообщение (SMPP deliver_sm).
type inboundMO struct {
	Phone       string // MSISDN абонента (sourceAddr)
	ShortNumber string // короткий номер (destinationAddr)
	Text        string // текст сообщения
}

func parseInboundMO(sms inboundSMSDTO) inboundMO {
	return inboundMO{
		Phone:       normalizePhone(string(sms.Phone)),
		ShortNumber: normalizeShortNumber(string(sms.ShortNumber)),
		Text:        strings.TrimSpace(sms.Text),
	}
}

// isActivationOne — общая точка входа сценариев 1 и 2: текст «1» на короткий номер 1020.
func (m inboundMO) isActivationOne() bool {
	return m.ShortNumber == activationShortNumber && strings.EqualFold(m.Text, "1")
}

// isStopOn1020 — вход сценария 5: текст «STOP» на номер 1020.
func (m inboundMO) isStopOn1020() bool {
	return m.ShortNumber == activationShortNumber && strings.EqualFold(m.Text, "STOP")
}

// isLanguageChangeOn1020 — вход сценария 7: текст «RUS», «ARM» или «ENG» на номер 1020.
func (m inboundMO) isLanguageChangeOn1020() bool {
	if m.ShortNumber != activationShortNumber {
		return false
	}
	switch strings.ToUpper(m.Text) {
	case "RUS", "ARM", "ENG":
		return true
	default:
		return false
	}
}

// languageCommandCode возвращает нормализованный код языка (ru, arm, en) для MO сценария 7.
func (m inboundMO) languageCommandCode() string {
	if !m.isLanguageChangeOn1020() {
		return ""
	}
	return normalizeLang(m.Text)
}

// isKnownCommandOn1020 — true для MO на 1020 с текстом 1, STOP, RUS, ARM или ENG.
func (m inboundMO) isKnownCommandOn1020() bool {
	return m.isActivationOne() || m.isStopOn1020() || m.isLanguageChangeOn1020()
}

// isUnknownCommandOn1020 — вход сценария 8: любой другой текст на номер 1020.
func (m inboundMO) isUnknownCommandOn1020() bool {
	return m.ShortNumber == activationShortNumber && !m.isKnownCommandOn1020()
}
