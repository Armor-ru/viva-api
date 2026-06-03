package viva_api

import (
	"fmt"
	"os"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *viva) InitHandlers() {
	if s.extTransport != nil {
		s.initMiddleWare()

		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", func(ctx types.HandlerContext) {
			s.webhookHandler(ctx, types.OrderTypeNew)
		})
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", s.handleExtAppPartnerProductActivation)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", func(ctx types.HandlerContext) {
			s.webhookHandler(ctx, types.OrderTypeCancel)
		})
		s.extTransport.Subscribe("POST /landing/init-subscription", s.landingInitHandler)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.landingConfirmHandler)
	}

	if s.intTransport != nil {
		if _, err := s.intTransport.Subscribe("order/completed", s.orderCompleteHandler); err != nil {
			logger.Error().Err(err).Str("topic", "order/completed").Msg("NATS subscribe failed")
		} else {
			logger.Info().
				Str("topic", "order/completed").
				Msg("subscribed on NATS for order/completed (SDS publishes after processing order/create)")
		}
		if _, err := s.intTransport.Subscribe(orderExpiresSubscribePath, s.orderExpiresHandler); err != nil {
			logger.Error().Err(err).Str("path", orderExpiresSubscribePath).Msg("NATS subscribe failed")
		} else {
			logger.Info().
				Str("path", orderExpiresSubscribePath).
				Str("natsSubject", "order.expires").
				Msg("subscribed on NATS for order.expires (SDS publishes one day before subscription end)")
		}
	}

	if err := s.loadCatalog(); err != nil {
		panic(err)
	}

	if s.ussdTransport != nil {
		if _, err := s.ussdTransport.Subscribe("smpp/inbound", s.ussdHandler); err != nil {
			logger.Error().Err(err).Msg("notify inbound subscribe failed")
		}
		if err := s.ussdTransport.Connect(); err != nil {
			logger.Error().Err(err).Msg("notify transport connect failed")
		}
	}
}

func (s *viva) initMiddleWare() {
	signature := httpTransport.Signature(httpTransport.SignatureConfig{
		Secrets: s.secrets,
		Header:  "X-Signature",
	})

	s.extTransport.Middleware(cors.New(cors.Config{
		Next: func(c *fiber.Ctx) bool {
			return !isLandingPath(c.Path())
		},
		AllowOrigins: "*",
		AllowMethods: "POST,OPTIONS",
		AllowHeaders: "Content-Type,Accept,X-MSISDN,X-Msisdn,X-Phone-Number",
	}))
	s.extTransport.Middleware(func(c *fiber.Ctx) error {
		if isLandingPath(c.Path()) {
			return c.Next()
		}
		return signature(c)
	})
}

func isLandingPath(path string) bool {
	return path == "/landing" || strings.HasPrefix(path, "/landing/")
}

func (s *viva) loadCatalog() error {
	dir := strings.TrimSpace(s.catalogDir)
	if dir == "" {
		dir = "catalog"
	}
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("stat catalog directory %q: %w", dir, err)
	}

	cat := catalog.NewCatalog()
	if err := cat.Load(dir); err != nil {
		return err
	}
	defaultLang := ""
	if s.langStore != nil {
		defaultLang = s.langStore.DefaultLang()
	}
	_ = cat.SetDefaultLang(defaultLang)
	s.catalog = cat
	logger.Info().Str("path", dir).Msg("catalog loaded")
	return nil
}
