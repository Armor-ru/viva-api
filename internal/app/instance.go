package viva_api

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type LangStore map[string]langEntry

type langEntry struct {
	lang      string
	expiresAt time.Time
}

const langPreferenceTTL = 90 * 24 * time.Hour

func (s *Viva) storedLang(phone string) string {
	phone = strings.TrimSpace(strings.TrimPrefix(phone, "+"))
	if phone == "" {
		return ""
	}
	e, ok := s.langStore[phone]
	if !ok || time.Now().After(e.expiresAt) {
		return ""
	}
	return strings.TrimSpace(e.lang)
}

func (s *Viva) storeLang(phone, lang string) {
	phone = strings.TrimSpace(strings.TrimPrefix(phone, "+"))
	lang = strings.TrimSpace(lang)
	if phone == "" || lang == "" {
		return
	}
	s.langStore[phone] = langEntry{
		lang:      lang,
		expiresAt: time.Now().Add(langPreferenceTTL),
	}
}

type UssdRequest struct {
	Phone       string `json:"sourceAddr" validate:"required"`
	ShortNumber string `json:"destinationAddr" validate:"required"`
	Text        string `json:"shortMessage" validate:"required"`
}

type ExtReq struct {
	PhoneNum    string `json:"phoneNum" validate:"required"`
	ProductCode string `json:"productCode" validate:"required"`
}

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
	}

	if s.ussdTransport != nil {
		if _, err := s.ussdTransport.Subscribe("smpp/inbound", s.ussdHandler); err != nil {
			logger.Error().Err(err).Msg("smpp inbound subscribe failed")
		}
	}
}

func (s *Viva) buildLandingConfirmURL(phone, productName, productCode, lang string) (string, error) {
	base := strings.TrimSpace(s.landingConfirmURL)
	if base == "" {
		return "", fmt.Errorf("landing confirm url is empty")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("landing confirm url: %w", err)
	}
	q := u.Query()
	q.Set("phone", strings.TrimPrefix(strings.TrimSpace(phone), "+"))
	q.Set("productName", strings.TrimSpace(productName))
	q.Set("productCode", strings.TrimSpace(productCode))
	if lang = strings.TrimSpace(lang); lang != "" {
		q.Set("lang", lang)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
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
func (s *Viva) ussd(req UssdRequest) {
	product, err := s.catalog.GetProductByShortNumber(req.ShortNumber)
	if err != nil {
		logAppError(err, "get product by short number failed")
		return
	}
	lang := s.storedLang(req.Phone)
	if lang == "" {
		lang = "ru"
	}
	switch req.Text {
	case "1":
		orderID := s.getOrderId(req.Phone, product.ExternalId)
		order, err := s.getOrder(orderID)
		exists := err == nil && strings.TrimSpace(order.ID) != ""

		data := map[string]interface{}{
			"Phone":       req.Phone,
			"ShortNumber": req.ShortNumber,
			"ExternalID":  product.ExternalId,
		}

		var key string

		switch {
		case exists && s.isActiveOrder(order):
			key = "already_active"

		case !exists:
			if _, err := s.landingInit(SubscriptionInitRequest{
				PhoneNum:    req.Phone,
				ProductName: product.ExternalId,
				Lang:        lang,
			}); err != nil {
				logAppError(err, "init subscription failed")
				return
			}

			landingURL, err := s.buildLandingConfirmURL(req.Phone, product.ExternalId, product.ExternalId, lang)
			if err != nil {
				logAppError(err, "build landing url failed")
				return
			}
			data["LandingURL"] = landingURL
			key = "otp_landing"

		default:
			logAppError(
				errs.WrapWithFields(
					fmt.Errorf("order exists but not active"),
					map[string]interface{}{
						"orderId":     orderID,
						"phone":       req.Phone,
						"shortNumber": req.ShortNumber,
						"externalId":  product.ExternalId,
					},
				),
				"skip init subscription",
			)
			return
		}
		if err := s.sendProductNotify(product, req.Phone, key, data, lang); err != nil {
			logAppError(err, "notify failed")
		}
	case "STOP":
		productCode := product.ExternalId
		orderID := s.getOrderId(req.Phone, productCode)
		order, err := s.getOrder(orderID)
		exists := err == nil && strings.TrimSpace(order.ID) != ""

		data := map[string]interface{}{
			"Phone":       req.Phone,
			"ShortNumber": req.ShortNumber,
			"ExternalID":  productCode,
		}
		key := "service_deactivated"
		if exists && !s.isActiveOrder(order) {
			key = "already_deactivated"
		} else if err := s.removeSubscription(req.Phone, productCode); err != nil {
			logAppError(err, "remove subscription failed")
			return
		}
		if err := s.sendProductNotify(product, req.Phone, key, data, lang); err != nil {
			logAppError(err, "notify failed")
		}
	case "RUS", "ENG", "ARM":
		switch req.Text {
		case "RUS":
			lang = "ru"
		case "ENG":
			lang = "en"
		case "ARM":
			lang = "arm"
		}
		s.storeLang(req.Phone, lang)
		if err := s.sendProductNotify(product, req.Phone, "language_changed", map[string]interface{}{
			"Phone": req.Phone, "ShortNumber": req.ShortNumber, "Language": lang,
		}, lang); err != nil {
			logAppError(err, "notify failed")
		}
	default:
		if err := s.sendProductNotify(product, req.Phone, "unknown_command", map[string]interface{}{
			"Phone": req.Phone, "ShortNumber": req.ShortNumber, "Text": req.Text,
		}, lang); err != nil {
			logAppError(err, "notify failed")
		}
	}
}

func (s *Viva) orderCompleteHandler(ctx types.HandlerContext) {
	var order Order
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.completed")
	if err := s.completeOrder(order); err != nil {
		logAppError(err, "complete order failed")
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
