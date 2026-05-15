package viva_api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Armor-ru/sds-go/pkg/logger"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/Armor-ru/viva-api/internal/vivaclient"
)

const requestTimeout = 25 * time.Second

var msisdnHeaderNames = []string{
	"X-MSISDN",
	"X-Msisdn",
	"X-Phone-Number",
}

type landingInitBody struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName"`
	SkipConfirm bool   `json:"skipConfirm"`
	ProductCode string `json:"productCode"`
	SmsScenario string `json:"smsScenario,omitempty"`
	Locale      string `json:"locale,omitempty"`
}

type landingConfirmBody struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName"`
	OTP         string `json:"otp" validate:"required"`
	ProductCode string `json:"productCode"`
	SmsScenario string `json:"smsScenario,omitempty"`
	Locale      string `json:"locale,omitempty"`
}

type landingSubscriberInfoBody struct {
	PhoneNum string `json:"phoneNum"`
	Locale   string `json:"locale,omitempty"`
}

func (s *Viva) landingInit(ctx types.HandlerContext)          { s.handleInit(ctx, "") }
func (s *Viva) landingInitLocalized(ctx types.HandlerContext) { s.handleInit(ctx, ctx.Param("locale")) }
func (s *Viva) landingConfirm(ctx types.HandlerContext)       { s.handleConfirm(ctx, "") }
func (s *Viva) landingConfirmLocalized(ctx types.HandlerContext) {
	s.handleConfirm(ctx, ctx.Param("locale"))
}
func (s *Viva) landingSubscriberGET(ctx types.HandlerContext) { s.handleSubscriberGET(ctx, "") }
func (s *Viva) landingSubscriberGETLocalized(ctx types.HandlerContext) {
	s.handleSubscriberGET(ctx, ctx.Param("locale"))
}
func (s *Viva) landingSubscriberPOST(ctx types.HandlerContext) { s.handleSubscriberPOST(ctx, "") }
func (s *Viva) landingSubscriberPOSTLocalized(ctx types.HandlerContext) {
	s.handleSubscriberPOST(ctx, ctx.Param("locale"))
}

func (s *Viva) handleInit(ctx types.HandlerContext, pathLocale string) {
	if !s.ensurePartner(ctx) {
		return
	}

	var body landingInitBody
	ctx.Data(&body)

	locale, err := landingPickLocale(pathLocale, body.Locale)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	product := strings.TrimSpace(body.ProductName)
	if product == "" {
		respondError(ctx, 400, "productName required")
		return
	}

	phone, _, err := extractMSISDN(ctx, body.PhoneNum)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	if !body.SkipConfirm {
		respondError(ctx, 400, "skipConfirm must be true; confirm requires OTP")
		return
	}

	c, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	initRes, err := s.vivaPartner.InitSubscription(c, phone, product)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva InitSubscription")
		respondError(ctx, 502, err.Error())
		return
	}

	respondOK(ctx, map[string]interface{}{"init": initRes, "locale": locale})
}

func (s *Viva) handleConfirm(ctx types.HandlerContext, pathLocale string) {
	if !s.ensurePartner(ctx) {
		return
	}

	var body landingConfirmBody
	ctx.Data(&body)

	locale, err := landingPickLocale(pathLocale, body.Locale)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	product := strings.TrimSpace(body.ProductName)
	if product == "" {
		respondError(ctx, 400, "productName required")
		return
	}

	otp := strings.TrimSpace(body.OTP)
	if otp == "" {
		respondError(ctx, 400, "otp required")
		return
	}

	phone, _, err := extractMSISDN(ctx, body.PhoneNum)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	c, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	res, err := s.vivaPartner.ConfirmSubscription(c, phone, product, &otp)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription (OTP)")
		respondError(ctx, 502, err.Error())
		return
	}

	if res.ResultCode == 7 {
		respondOK(ctx, map[string]interface{}{"confirm": res, "locale": locale})
		return
	}

	if res.ResultCode != 0 {
		respondError(ctx, 502, vivaErrorMessage(res))
		return
	}

	productCode := strings.TrimSpace(body.ProductCode)
	if productCode == "" {
		respondError(ctx, 400, "productCode required")
		return
	}
	s.sendOrderCreate(types.OrderTypeNew, phone, productCode, body.SmsScenario, locale)
	respondOK(ctx, map[string]interface{}{"confirm": res, "locale": locale})
}

func (s *Viva) handleSubscriberGET(ctx types.HandlerContext, pathLocale string) {
	if _, err := landingPickLocale(pathLocale, ""); err != nil {
		respondError(ctx, 400, err.Error())
		return
	}
	if !s.ensurePartner(ctx) {
		return
	}

	phone := cleanMSISDN(ctx.Param("phoneNum"))
	if phone == "" {
		respondError(ctx, 400, "phoneNum required")
		return
	}
	s.fetchSubscriberInfo(ctx, phone)
}

func (s *Viva) handleSubscriberPOST(ctx types.HandlerContext, pathLocale string) {
	if !s.ensurePartner(ctx) {
		return
	}

	var body landingSubscriberInfoBody
	ctx.Data(&body)

	if _, err := landingPickLocale(pathLocale, body.Locale); err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	phone, _, err := extractMSISDN(ctx, body.PhoneNum)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}
	s.fetchSubscriberInfo(ctx, phone)
}

func (s *Viva) fetchSubscriberInfo(ctx types.HandlerContext, phone string) {
	c, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	info, err := s.vivaPartner.GetSubscriberInfo(c, phone)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			respondError(ctx, 404, err.Error())
			return
		}
		logger.Error().Err(err).Str("phone", phone).Msg("Viva GetSubscriberInfo")
		respondError(ctx, 502, err.Error())
		return
	}
	respondOK(ctx, info)
}

func (s *Viva) ensurePartner(ctx types.HandlerContext) bool {
	if s.vivaPartner == nil {
		respondError(ctx, 503, "viva api not configured")
		return false
	}
	return true
}

func extractMSISDN(ctx types.HandlerContext, bodyPhone string) (string, bool, error) {
	var headers map[string]string
	ctx.Headers(&headers)

	for key, value := range headers {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		for _, name := range msisdnHeaderNames {
			if strings.EqualFold(strings.TrimSpace(key), name) {
				return cleanMSISDN(value), true, nil
			}
		}
	}

	if p := cleanMSISDN(bodyPhone); p != "" {
		return p, false, nil
	}
	return "", false, fmt.Errorf("phoneNum missing: provide in body or via MSISDN header")
}

func cleanMSISDN(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimPrefix(s, "+")
}

func respondOK(ctx types.HandlerContext, data interface{}) {
	_ = ctx.Response(data)
}

func respondError(ctx types.HandlerContext, status int, msg string) {
	_ = ctx.Response(httpt.MsgResponse{
		Status:  status,
		Payload: map[string]string{"error": msg},
	})
}

func vivaErrorMessage(res *vivaclient.ResponseModel) string {
	if res == nil {
		return "viva: empty response body"
	}
	return fmt.Sprintf("viva subscription resultCode=%d message=%v", res.ResultCode, res.Message)
}
