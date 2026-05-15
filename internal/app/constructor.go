package viva_api

import (
	"context"

	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/Armor-ru/viva-api/internal/vivaclient"
)

type PartnerSubscriptionAPI interface {
	GetSubscriberInfo(ctx context.Context, msisdn string) (*vivaclient.GetSubInfoResponse, error)
	InitSubscription(ctx context.Context, phoneNum, productName string) (*vivaclient.ResponseModel, error)
	ConfirmSubscription(ctx context.Context, phoneNum, productName string, otp *string) (*vivaclient.ResponseModel, error)
}

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

func WithAccountId(accountId string) func(*Viva) {
	return func(s *Viva) {
		s.accountId = accountId
	}
}

func WithVivaPartner(api PartnerSubscriptionAPI) func(*Viva) {
	return func(s *Viva) {
		s.vivaPartner = api
	}
}
