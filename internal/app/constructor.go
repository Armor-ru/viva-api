package viva_api

import (
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/Armor-ru/viva-api/internal/app/utils"
)

func New(options ...func(*Viva)) Viva {
	instance := Viva{}
	for _, option := range options {
		option(&instance)
	}
	instance.InitSMSInfrastructure()
	return instance
}

func WithIntTransport(transport types.Transport) func(*Viva) {
	return func(s *Viva) {
		s.IntTransport = transport
	}
}

func WithExtTransport(transport types.Transport) func(*Viva) {
	return func(s *Viva) {
		s.ExtTransport = transport
	}
}

func WithSecrets(secrets []string) func(*Viva) {
	return func(s *Viva) {
		s.Secrets = secrets
	}
}

func WithSmppConfig(config utils.SmppConfig) func(*Viva) {
	return func(s *Viva) {
		s.Smpp = config
	}
}

func WithAccountId(id string) func(*Viva) {
	return func(s *Viva) {
		s.AccountId = id
	}
}

func WithVivaPartner(api PartnerSubscriptionAPI) func(*Viva) {
	return func(s *Viva) {
		s.VivaPartner = api
	}
}
