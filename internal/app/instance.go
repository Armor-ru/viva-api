package viva_api

import (
	"strings"
	"text/template"

	"github.com/Armor-ru/sds-go/pkg/logger"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Viva struct {
	intTransport types.Transport
	extTransport types.Transport

	secrets     []string
	smpp        SmppConfig
	accountId   string
	vivaPartner PartnerSubscriptionAPI

	testTariffs []string
	channels    Channels

	activationTpl *template.Template
	scenarioTpl   map[string]map[string]*template.Template
	smppSender    *SmppSender
}

func (s *Viva) InitHandlers() {
	s.initSMS()

	if s.extTransport != nil {
		s.initMiddleware()

		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", s.extActivationRequest)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", s.extActivation)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", s.extRemove)

		s.extTransport.Subscribe("POST /landing/init-subscription", s.landingInit)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.landingConfirm)
		s.extTransport.Subscribe("GET /landing/subscriber-info/:phoneNum", s.landingSubscriberGET)
		s.extTransport.Subscribe("POST /landing/subscriber-info", s.landingSubscriberPOST)

		s.extTransport.Subscribe("POST /landing/:locale/init-subscription", s.landingInitLocalized)
		s.extTransport.Subscribe("POST /landing/:locale/confirm-subscription", s.landingConfirmLocalized)
		s.extTransport.Subscribe("GET /landing/:locale/subscriber-info/:phoneNum", s.landingSubscriberGETLocalized)
		s.extTransport.Subscribe("POST /landing/:locale/subscriber-info", s.landingSubscriberPOSTLocalized)
	}

	if s.intTransport != nil {
		s.intTransport.Subscribe("order/completed", s.onOrderCompleted)
	}
}

func (s *Viva) initMiddleware() {
	ht, ok := s.extTransport.(*httpt.Transport)
	sig := httpt.Signature(httpt.SignatureConfig{Secrets: s.secrets, Header: "X-Signature"})
	if !ok {
		logger.Warn().Msg("extTransport is not *http.Transport; webhook signature middleware skipped")
		return
	}
	ht.Middleware(cors.New(cors.Config{
		Next: func(c *fiber.Ctx) bool {
			return !strings.HasPrefix(c.Path(), "/landing/")
		},
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Accept",
	}))
	ht.Middleware(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/landing/") {
			return c.Next()
		}
		return sig(c)
	})
}
