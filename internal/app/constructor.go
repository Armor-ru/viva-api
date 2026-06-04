package viva_api

import (
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
)

type Option func(*Viva)

func New(options ...Option) *Viva {
	instance := &Viva{
		catalog:   NewCatalog(),
		langStore: make(LangStore),
	}

	for _, option := range options {
		option(instance)
	}

	instance.InitHandlers()

	return instance
}

func WithIntTransport(transport types.Transport) Option {
	return func(s *Viva) {
		s.intTransport = transport
	}
}

func WithExtTransport(transport types.Transport) Option {
	return func(s *Viva) {
		s.extTransport = transport
	}
}

func WithUssdTransport(transport types.Transport) Option {
	return func(s *Viva) {
		s.ussdTransport = transport
	}
}

func WithSecrets(secrets []string) Option {
	return func(s *Viva) {
		s.secrets = secrets
	}
}

func WithCatalogDir(dir string) Option {
	return func(s *Viva) {
		s.catalogDir = dir
	}
}

func WithAccountId(accountId string) Option {
	return func(s *Viva) {
		s.accountId = accountId
	}
}

func WithVivaClient(client *vivaclient.Client) Option {
	return func(s *Viva) {
		s.vivaClient = client
	}
}

func WithLandingConfirmURL(url string) Option {
	return func(s *Viva) {
		s.landingConfirmURL = strings.TrimSpace(url)
	}
}
