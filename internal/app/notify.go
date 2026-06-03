package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func (s *Viva) notify(phone, text string) error {
	if s.ussdTransport == nil {
		return fmt.Errorf("ussdTransport is not configured")
	}
	phone = strings.TrimSpace(strings.TrimPrefix(phone, "+"))
	text = strings.TrimSpace(text)
	if phone == "" || text == "" {
		return fmt.Errorf("phone and text are required")
	}
	_, err := s.ussdTransport.Send("", map[string]interface{}{
		"to":   phone,
		"text": text,
	}, types.SendOptions{})
	return err
}
