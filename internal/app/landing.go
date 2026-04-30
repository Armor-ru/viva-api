package viva_api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Armor-ru/sds-go/pkg/logger"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/Armor-ru/sds-go/pkg/types"
)

// Имена заголовков с MSISDN после обогащения на стороне Viva (вариант A). Регистр не важен.
var landingMSISDNHeaderNames = []string{
	"X-MSISDN",
	"X-Msisdn",
	"X-Phone-Number",
}

type LandingInitBody struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName"`
	SkipConfirm bool   `json:"skipConfirm"`
	// ProductCode — UUID продукта для ЗК (как во вебхуке); если пусто — из конфига vivaApi.orderProductCode.
	ProductCode string `json:"productCode"`
	SmsScenario string `json:"smsScenario,omitempty"`
}

type LandingConfirmBody struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName"`
	OTP         string `json:"otp" validate:"required"`
	ProductCode string `json:"productCode"`
	SmsScenario string `json:"smsScenario,omitempty"`
}

func (s *Viva) LandingInitSubscriptionHandler(ctx types.HandlerContext) {
	if !s.ensureVivaPartner(ctx) {
		return
	}
	var body LandingInitBody
	ctx.Data(&body)

	prod, ok := s.landingProduct(body.ProductName)
	if !ok {
		landingErr(ctx, 400, "productName required")
		return
	}
	phone, err := landingResolveMSISDN(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	c, cancel := landingCtx()
	defer cancel()

	initRes, err := s.vivaPartner.InitSubscription(c, phone, prod)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva InitSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	if body.SkipConfirm {
		_ = ctx.Response(map[string]any{"init": initRes})
		return
	}

	confirmRes, err := s.vivaPartner.ConfirmSubscription(c, phone, prod, nil)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	s.landingEmitNewOrder(phone, body.ProductCode, body.SmsScenario)
	_ = ctx.Response(map[string]any{"init": initRes, "confirm": confirmRes})
}

func (s *Viva) LandingConfirmSubscriptionHandler(ctx types.HandlerContext) {
	if !s.ensureVivaPartner(ctx) {
		return
	}
	var body LandingConfirmBody
	ctx.Data(&body)

	prod, ok := s.landingProduct(body.ProductName)
	if !ok {
		landingErr(ctx, 400, "productName required")
		return
	}
	otp := strings.TrimSpace(body.OTP)
	if otp == "" {
		landingErr(ctx, 400, "otp required")
		return
	}
	phone, err := landingResolveMSISDN(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	c, cancel := landingCtx()
	defer cancel()

	res, err := s.vivaPartner.ConfirmSubscription(c, phone, prod, &otp)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription (OTP)")
		landingErr(ctx, 502, err.Error())
		return
	}
	s.landingEmitNewOrder(phone, body.ProductCode, body.SmsScenario)
	_ = ctx.Response(map[string]any{"confirm": res})
}

func (s *Viva) LandingGetSubscriberInfoGETHandler(ctx types.HandlerContext) {
	if !s.ensureVivaPartner(ctx) {
		return
	}
	phone := normalizeMSISDN(ctx.Param("phoneNum"))
	if phone == "" {
		landingErr(ctx, 400, "phoneNum required")
		return
	}
	s.landingRespondSubscriberInfo(ctx, phone)
}

type landingSubscriberInfoBody struct {
	PhoneNum string `json:"phoneNum"`
}

func (s *Viva) LandingGetSubscriberInfoPOSTHandler(ctx types.HandlerContext) {
	if !s.ensureVivaPartner(ctx) {
		return
	}
	var body landingSubscriberInfoBody
	ctx.Data(&body)
	phone, err := landingResolveMSISDN(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	s.landingRespondSubscriberInfo(ctx, phone)
}

func (s *Viva) landingRespondSubscriberInfo(ctx types.HandlerContext, phone string) {
	c, cancel := landingCtx()
	defer cancel()
	info, err := s.vivaPartner.GetSubscriberInfo(c, phone)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			landingErr(ctx, 404, err.Error())
			return
		}
		logger.Error().Err(err).Str("phone", phone).Msg("Viva GetSubscriberInfo")
		landingErr(ctx, 502, err.Error())
		return
	}
	_ = ctx.Response(info)
}

func (s *Viva) ensureVivaPartner(ctx types.HandlerContext) bool {
	if s.vivaPartner != nil {
		return true
	}
	landingErr(ctx, 503, "viva api not configured")
	return false
}

// landingProduct — имя продукта из тела или default из конфига.
func (s *Viva) landingProduct(name string) (string, bool) {
	p := strings.TrimSpace(name)
	if p == "" {
		p = s.defaultProductName
	}
	if p == "" {
		return "", false
	}
	return p, true
}

func landingCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 25*time.Second)
}

func landingErr(ctx types.HandlerContext, status int, msg string) {
	_ = ctx.Response(httpt.MsgResponse{
		Status:  status,
		Payload: map[string]string{"error": msg},
	})
}

func normalizeMSISDN(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	return s
}

func landingResolveMSISDN(ctx types.HandlerContext, bodyPhone string) (string, error) {
	var hdr map[string]string
	ctx.Headers(&hdr)
	for k, v := range hdr {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		for _, name := range landingMSISDNHeaderNames {
			if strings.EqualFold(strings.TrimSpace(k), name) {
				return normalizeMSISDN(v), nil
			}
		}
	}
	if p := normalizeMSISDN(bodyPhone); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("phoneNum missing: set JSON phoneNum (manual entry) or MSISDN header e.g. X-MSISDN (Viva header enrichment)")
}

func (s *Viva) landingEmitNewOrder(phone, bodyProductCode, smsScenario string) {
	pc := strings.TrimSpace(bodyProductCode)
	if pc == "" {
		pc = strings.TrimSpace(s.orderProductCode)
	}
	s.sendOrderCreate(types.OrderTypeNew, phone, pc, smsScenario)
}
