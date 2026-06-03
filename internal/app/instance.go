package viva_api

import (
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type LangStore map[string]string

type UssdRequest struct {
	Phone       string `json:"sourceAddr" validate:"required"`
	ShortNumber string `json:"destinationAddr" validate:"required"`
	Text        string `json:"shortMessage" validate:"required"`
}

type Viva struct {
	intTransport  types.Transport
	extTransport  types.Transport
	ussdTransport types.Transport
	vivaClient    *vivaclient.Client
	langStore     LangStore

	catalogDir string
	catalog    *Catalog
	secrets    []string
	accountId  string
}

func (s *Viva) InitHandlers() {
	if s.catalogDir != "" {
		if err := s.catalog.Load(s.catalogDir); err != nil {
			panic(err)
		}
	}

	if s.extTransport != nil {
		s.initMiddleWare()

		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", func(ctx types.HandlerContext) {
			s.webhookHandler(ctx, types.OrderTypeNew)
		})
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", func(ctx types.HandlerContext) {
			s.webhookHandler(ctx, types.OrderTypeRenew)
		})
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", func(ctx types.HandlerContext) {
			s.webhookHandler(ctx, types.OrderTypeCancel)
		})
		s.extTransport.Subscribe("POST /landing/init-subscription", s.landingInitHandler)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.landingConfirmHandler)
	}

	if s.intTransport != nil {
		s.intTransport.Subscribe("order/completed", s.orderCompleteHandler)
	}

	if s.ussdTransport != nil {
		if _, err := s.ussdTransport.Subscribe("smpp/inbound", s.ussdHandler); err != nil {
			logger.Error().Err(err).Msg("smpp inbound subscribe failed")
		}
	}
}

func (s *Viva) ussdHandler(ctx types.HandlerContext) {
	var req UssdRequest
	ctx.Data(&req)
	_ = req
}

func (s *Viva) orderCompleteHandler(ctx types.HandlerContext) {
	var order Order
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.completed")
	if err := s.completeOrder(order); err != nil {
		logger.Error().Err(err).Str("orderId", order.ID).Msg("complete order failed")
	}
}

func (s *Viva) webhookHandler(ctx types.HandlerContext, orderType types.OrderType) {
	var req ExtReq
	ctx.Data(&req)

	if err := s.createOrder(orderType, req.PhoneNum, req.ProductCode, ""); err != nil {
		logger.Error().Msg("can not create order, " + err.Error())
		return
	}

	ctx.Response("")
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
