package viva_api

import (
	"fmt"
	"strings"
	"time"

	transportSmpp "git.dev.armlab.pro/armor/sds-go/pkg/transport/smpp"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

const (
	notifyMaxAttempts = 5
	notifyRetryWait   = 500 * time.Millisecond
)

type smppConnector interface {
	Connect() error
	Connected() bool
}

func (s *Viva) notify(phone, text string) error {
	if s.notifyTransport == nil {
		return fmt.Errorf("notify transport is not configured")
	}

	msg := transportSmpp.Message{To: phone, Text: text}
	opts := types.SendOptions{Timeout: 5 * time.Second}

	var lastErr error
	for attempt := 0; attempt < notifyMaxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(notifyRetryWait)
			if conn, ok := s.notifyTransport.(smppConnector); ok && !conn.Connected() {
				_ = conn.Connect()
			}
		}

		_, lastErr = s.notifyTransport.Send("", msg, opts)
		if lastErr == nil {
			return nil
		}
		if !isRetryableNotifyErr(lastErr) {
			return lastErr
		}
	}
	return lastErr
}

func isRetryableNotifyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not connected") ||
		strings.Contains(msg, "disconnected") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "timeout")
}
