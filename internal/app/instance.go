package viva_api

import (
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Viva struct {
	intTransport  types.Transport
	extTransport  types.Transport
	ussdTransport types.Transport
	vivaClient    *vivaclient.Client
	langStore     LangStore

	catalogDir        string
	catalog           *Catalog
	secrets           []string
	accountId         string
	landingConfirmURL string
}

func (s *Viva) InitHandlers() {
	if s.catalogDir != "" {
		if err := s.catalog.Load(s.catalogDir); err != nil {
			logger.Fatal().Fields(errs.Fields(err)).Err(err).Msg("catalog load failed")
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
		s.intTransport.Subscribe("order/expires", s.orderExpiresHandler)
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
	req.Phone = strings.TrimSpace(strings.TrimPrefix(req.Phone, "+"))
	req.ShortNumber = strings.TrimSpace(req.ShortNumber)
	req.Text = strings.TrimSpace(strings.ToUpper(req.Text))
	logger.Info().
		Str("phone", req.Phone).
		Str("shortNumber", req.ShortNumber).
		Str("text", req.Text).
		Msg("smpp mo received")
	s.ussd(req)
}

func (s *Viva) orderCompleteHandler(ctx types.HandlerContext) {
	var order Order
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.completed")
	if err := s.completeOrder(order); err != nil {
		logAppError(err, "complete order failed")
	}
}

func (s *Viva) orderExpiresHandler(ctx types.HandlerContext) {
	var order Order
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.expires")
	if err := s.expireOrder(order); err != nil {
		logAppError(err, "expire order failed")
	}
}

func (s *Viva) webhookHandler(ctx types.HandlerContext, orderType types.OrderType) {
	var req ExtReq
	ctx.Data(&req)

	if err := s.createOrder(orderType, req.PhoneNum, req.ProductCode, ""); err != nil {
		logAppError(err, "create order failed")
		return
	}

	_ = ctx.Response("")
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
