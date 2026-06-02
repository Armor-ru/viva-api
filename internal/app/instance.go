package viva_api

import (
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Viva struct {
	intTransport types.Transport
	extTransport types.Transport

	secrets []string
	sms     SmsConfig

	catalog pipeline.Catalog
	engine  *pipeline.Engine

	testTariffs []string
	channels    Channels

	notifyTransport types.Transport

	accountId  string
	vivaClient *vivaclient.Client
}

func (s *Viva) InitHandlers() {
	if err := s.loadCatalog(); err != nil {
		panic(err)
	}

	if s.extTransport != nil {
		s.initMiddleWare()

		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", s.ExtAppPartnerProductActivationRequestHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", s.ExtAppPartnerProductActivationHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", s.ExtAppPartnerProductRemoveHandler)
		s.extTransport.Subscribe("POST /landing/init-subscription", s.LandingInitHandler)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.LandingConfirmHandler)
	}

	if s.intTransport != nil {
		s.intTransport.Subscribe("order/completed", s.onCompletedHandler)
		s.intTransport.Subscribe("order/expires", s.onExpiresHandler)
	}

	if s.notifyTransport == nil {
		return
	}
	if _, err := s.notifyTransport.Subscribe("smpp/inbound", s.handleInboundSMS); err != nil {
		logger.Error().Err(err).Msg("smpp inbound subscribe failed")
	}
	if err := s.notifyTransport.Connect(); err != nil {
		logger.Error().Err(err).Msg("notify transport connect failed")
	}
}

func (s *Viva) initMiddleWare() {
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
