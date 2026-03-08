package viva_api

import "github.com/Armor-ru/sds-go/pkg/types"

func New(options ...func(*Viva)) Viva {
	instance := Viva{}
	for _, option := range options {
		option(&instance)
	}
	instance.InitHandlers()
	return instance
}

func WithIntTransport(transport types.Transport) func(*Viva) {
	return func(s *Viva) {
		s.intTransport = transport
	}
}

func WithExtTransport(transport types.Transport) func(*Viva) {
	return func(s *Viva) {
		s.extTransport = transport
	}
}

func WithSecrets(secrets []string) func(*Viva) {
	return func(s *Viva) {
		s.secrets = secrets
	}
}

func WithSmppConfig(config SmppConfig) func(*Viva) {
	return func(s *Viva) {
		s.smpp = config
	}
}
