package viva_api

import (
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
)

func New(options ...func(*viva)) Viva {
	instance := viva{}
	for _, option := range options {
		option(&instance)
	}
	instance.InitHandlers()
	return &instance
}

func WithIntTransport(transport types.Transport) func(*viva) {
	return func(s *viva) { s.intTransport = transport }
}

func WithExtTransport(transport types.Transport) func(*viva) {
	return func(s *viva) { s.extTransport = transport }
}

func WithUssdTransport(transport types.Transport) func(*viva) {
	return func(s *viva) { s.ussdTransport = transport }
}

func WithSecrets(secrets []string) func(*viva) {
	return func(s *viva) { s.secrets = secrets }
}

func WithCatalogDir(dir string) func(*viva) {
	return func(s *viva) { s.catalogDir = dir }
}

func WithDefaultLanguage(lang string, ttl ...time.Duration) func(*viva) {
	return func(s *viva) { s.langStore = NewLangStore(lang, ttl...) }
}

func WithAccountId(accountId string) func(*viva) {
	return func(s *viva) { s.accountId = accountId }
}

func WithVivaClient(client *vivaclient.Client) func(*viva) {
	return func(s *viva) { s.vivaClient = client }
}

func WithLandingConfirmURL(url string) func(*viva) {
	return func(s *viva) { s.landingConfirmURL = strings.TrimSpace(url) }
}
