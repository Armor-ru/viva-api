package viva

import (
	"fmt"
	"time"

	viva_api "git.dev.armlab.pro/armor/viva-api/internal/app"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
)

// SmppSender отправляет MT через прямой SMPP bind (только интеграционные тесты).
type SmppSender struct {
	cfg viva_api.SmppConfig
}

func NewSmppSender(cfg viva_api.SmppConfig) *SmppSender {
	return &SmppSender{cfg: cfg}
}

func (s *SmppSender) Send(msisdn, message string) error {
	if len(s.cfg.Endpoint) == 0 || s.cfg.Endpoint[0] == "" {
		return fmt.Errorf("smpp endpoint is empty")
	}
	if s.cfg.Auth.User == "" || s.cfg.Auth.Password == "" {
		return fmt.Errorf("smpp auth user or password is empty")
	}

	addr := s.cfg.Address
	tx := &smpp.Transceiver{
		Addr:        s.cfg.Endpoint[0],
		User:        s.cfg.Auth.User,
		Passwd:      s.cfg.Auth.Password,
		RespTimeout: 5 * time.Second,
	}

	conn := tx.Bind()
	select {
	case st := <-conn:
		if st.Error() != nil {
			return fmt.Errorf("smpp bind failed: %w", st.Error())
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("smpp bind timeout")
	}

	_, err := tx.Submit(&smpp.ShortMessage{
		Src:      addr.SourceAddr,
		Dst:      msisdn,
		Text:     pdutext.Raw(message),
		Register: 0,
	})
	_ = tx.Close()
	return err
}
