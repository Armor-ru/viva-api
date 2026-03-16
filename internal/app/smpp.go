package viva_api

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
)

type SmppSender struct {
	cfg SmppConfig
}

func NewSmppSender(cfg SmppConfig) *SmppSender {
	return &SmppSender{cfg: cfg}
}

func (s *SmppSender) Send(msisdn, message string) (err error) {
	if len(s.cfg.Endpoint) == 0 {
		return errors.New("smpp endpoint is empty")
	}
	if s.cfg.Auth.User == "" || s.cfg.Auth.Password == "" {
		return errors.New("smpp auth user or password is empty")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	endpoint := s.cfg.Endpoint[r.Intn(len(s.cfg.Endpoint))]

	// дополняем ошибку 
	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"smpp send failed: endpoint=%s msisdn=%s message_len=%d: %w",
				endpoint,
				msisdn,
				len(message),
				err,
			)
		}
	}()

	tx := &smpp.Transceiver{
		Addr:   endpoint,
		User:   s.cfg.Auth.User,
		Passwd: s.cfg.Auth.Password,
	}

	conn := tx.Bind()
	defer tx.Close()

	select {
	case st := <-conn:
		if st.Error() != nil {
			return st.Error()
		}
	case <-time.After(5 * time.Second):
		return errors.New("smpp bind timeout")
	}

	_, err = tx.Submit(&smpp.ShortMessage{
		Dst:  msisdn,
		Text: pdutext.Raw(message),
	})

	return
}
