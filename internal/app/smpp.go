package viva_api

import (
	"math/rand"
	"time"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
)

// Error
type SmppError struct {
	Message string
	Err     error
	Fields  map[string]interface{}
}

func (e *SmppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *SmppError) Unwrap() error {
	return e.Err
}

// SMPP
type SmppSender struct {
	cfg SmppConfig
}

func NewSmppSender(cfg SmppConfig) *SmppSender {
	return &SmppSender{cfg: cfg}
}

func (s *SmppSender) Send(msisdn, message string) error {
	if len(s.cfg.Endpoint) == 0 {
		return &SmppError{
			Message: "smpp endpoint is empty",
			Fields: map[string]interface{}{
				"msisdn":      msisdn,
				"message_len": len(message),
			},
		}
	}

	if s.cfg.Auth.User == "" || s.cfg.Auth.Password == "" {
		return &SmppError{
			Message: "smpp auth user or password is empty",
			Fields: map[string]interface{}{
				"msisdn":      msisdn,
				"message_len": len(message),
			},
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	endpoint := s.cfg.Endpoint[r.Intn(len(s.cfg.Endpoint))]

	fields := map[string]interface{}{
		"endpoint":    endpoint,
		"msisdn":      msisdn,
		"message_len": len(message),
	}

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
			return &SmppError{
				Message: "smpp bind failed",
				Err:     st.Error(),
				Fields:  fields,
			}
		}
	case <-time.After(5 * time.Second):
		return &SmppError{
			Message: "smpp bind timeout",
			Fields:  fields,
		}
	}

	_, err := tx.Submit(&smpp.ShortMessage{
		Dst:  msisdn,
		Text: pdutext.Raw(message),
	})
	if err != nil {
		return &SmppError{
			Message: "smpp submit failed",
			Err:     err,
			Fields:  fields,
		}
	}

	return nil
}
