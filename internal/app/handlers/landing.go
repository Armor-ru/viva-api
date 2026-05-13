package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
	"github.com/Armor-ru/viva-api/internal/app/service"
	"github.com/Armor-ru/viva-api/internal/app/utils"
	"github.com/Armor-ru/viva-api/internal/vivaclient"

	"github.com/Armor-ru/sds-go/pkg/logger"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/Armor-ru/sds-go/pkg/types"
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

func landingInitSubscriptionHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleInit(v, ctx, "")
	}
}

func landingInitSubscriptionLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleInit(v, ctx, ctx.Param("locale"))
	}
}

func landingConfirmSubscriptionHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleConfirm(v, ctx, "")
	}
}

func landingConfirmSubscriptionLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleConfirm(v, ctx, ctx.Param("locale"))
	}
}

func landingGetSubscriberInfoGETHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleSubscriberGET(v, ctx, "")
	}
}

func landingGetSubscriberInfoGETLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleSubscriberGET(v, ctx, ctx.Param("locale"))
	}
}

func landingGetSubscriberInfoPOSTHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleSubscriberPOST(v, ctx, "")
	}
}

func landingGetSubscriberInfoPOSTLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleSubscriberPOST(v, ctx, ctx.Param("locale"))
	}
}

func handleInit(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if !ensurePartner(v, ctx) {
		return
	}

	var body landingInitBody
	ctx.Data(&body)

	locale, err := utils.LandingPickLocale(pathLocale, body.Locale)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	product := strings.TrimSpace(body.ProductName)
	if product == "" {
		respondError(ctx, 400, "productName required")
		return
	}

	phone, fromHeader, err := extractMSISDN(ctx, body.PhoneNum)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	// Security hardening: do not allow OTP-less confirmation via Header Enrichment.
	// Always require explicit OTP confirmation via /landing/confirm-subscription.
	_ = fromHeader
	if !body.SkipConfirm {
		respondError(ctx, 400, "skipConfirm must be true; confirm requires OTP")
		return
	}

	c, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	initRes, err := v.VivaPartner.InitSubscription(c, phone, product)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva InitSubscription")
		respondError(ctx, 502, err.Error())
		return
	}

	respondOK(ctx, map[string]interface{}{"init": initRes, "locale": locale})
}

func handleConfirm(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if !ensurePartner(v, ctx) {
		return
	}

	var body landingConfirmBody
	ctx.Data(&body)

	locale, err := utils.LandingPickLocale(pathLocale, body.Locale)
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

	res, err := v.VivaPartner.ConfirmSubscription(c, phone, product, &otp)
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
	service.SendOrderCreate(v, types.OrderTypeNew, phone, productCode, body.SmsScenario, locale)
	respondOK(ctx, map[string]interface{}{"confirm": res, "locale": locale})
}

func handleSubscriberGET(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if _, err := utils.LandingPickLocale(pathLocale, ""); err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	if !ensurePartner(v, ctx) {
		return
	}

	phone := cleanMSISDN(ctx.Param("phoneNum"))
	if phone == "" {
		respondError(ctx, 400, "phoneNum required")
		return
	}

	fetchSubscriberInfo(ctx, v.VivaPartner, phone)
}

func handleSubscriberPOST(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if !ensurePartner(v, ctx) {
		return
	}

	var body landingSubscriberInfoBody
	ctx.Data(&body)

	if _, err := utils.LandingPickLocale(pathLocale, body.Locale); err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	phone, _, err := extractMSISDN(ctx, body.PhoneNum)
	if err != nil {
		respondError(ctx, 400, err.Error())
		return
	}

	fetchSubscriberInfo(ctx, v.VivaPartner, phone)
}

func fetchSubscriberInfo(ctx types.HandlerContext, client viva_api.PartnerSubscriptionAPI, phone string) {
	c, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	info, err := client.GetSubscriberInfo(c, phone)
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

func ensurePartner(v *viva_api.Viva, ctx types.HandlerContext) bool {
	if v.VivaPartner == nil {
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
