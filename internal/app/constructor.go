package viva_api

import (
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
)

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

func WithSmppTransport(transport types.Transport) func(*Viva) {
	return func(s *Viva) {
		s.notifyTransport = transport
	}
}

func WithSmsConfig(config SmsConfig) func(*Viva) {
	return func(s *Viva) {
		s.sms = config
	}
}

func WithAccountId(accountId string) func(*Viva) {
	return func(s *Viva) {
		s.accountId = accountId
	}
}

func WithVivaClient(client *vivaclient.Client) func(*Viva) {
	return func(s *Viva) {
		s.vivaClient = client
	}
}
