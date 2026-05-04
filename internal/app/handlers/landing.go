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

var landingMSISDNHeaderNames = []string{
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

func landingInitSubscriptionHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingInit(v, ctx, "")
	}
}

func landingInitSubscriptionLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingInit(v, ctx, ctx.Param("locale"))
	}
}

func runLandingInit(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if !ensureVivaPartner(v, ctx) {
		return
	}
	var body landingInitBody
	ctx.Data(&body)

	locale, err := utils.LandingPickLocale(pathLocale, body.Locale)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	prod, ok := landingProduct(v, body.ProductName)
	if !ok {
		landingErr(ctx, 400, "productName required")
		return
	}
	phone, msisdnFromHeader, err := landingResolveMSISDN(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	if !msisdnFromHeader && !body.SkipConfirm {
		landingErr(ctx, 400, "without carrier MSISDN header (e.g. X-MSISDN), skipConfirm must be true; complete flow with POST .../confirm-subscription and otp")
		return
	}

	c, cancel := landingCtx()
	defer cancel()

	initRes, err := v.VivaPartner.InitSubscription(c, phone, prod)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva InitSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	if body.SkipConfirm {
		_ = ctx.Response(map[string]any{"init": initRes, "locale": locale})
		return
	}

	confirmRes, err := v.VivaPartner.ConfirmSubscription(c, phone, prod, nil)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	if confirmRes.ResultCode == 7 {
		_ = ctx.Response(map[string]any{"init": initRes, "confirm": confirmRes, "locale": locale})
		return
	}
	if confirmRes.ResultCode != 0 {
		landingErr(ctx, 502, landingVivaResultMsg(confirmRes))
		return
	}
	landingEmitNewOrder(v, phone, body.ProductCode, body.SmsScenario, locale)
	_ = ctx.Response(map[string]any{"init": initRes, "confirm": confirmRes, "locale": locale})
}

func landingConfirmSubscriptionHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingConfirm(v, ctx, "")
	}
}

func landingConfirmSubscriptionLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingConfirm(v, ctx, ctx.Param("locale"))
	}
}

func runLandingConfirm(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if !ensureVivaPartner(v, ctx) {
		return
	}
	var body landingConfirmBody
	ctx.Data(&body)

	locale, err := utils.LandingPickLocale(pathLocale, body.Locale)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	prod, ok := landingProduct(v, body.ProductName)
	if !ok {
		landingErr(ctx, 400, "productName required")
		return
	}
	otp := strings.TrimSpace(body.OTP)
	if otp == "" {
		landingErr(ctx, 400, "otp required")
		return
	}
	phone, _, err := landingResolveMSISDN(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	c, cancel := landingCtx()
	defer cancel()

	res, err := v.VivaPartner.ConfirmSubscription(c, phone, prod, &otp)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription (OTP)")
		landingErr(ctx, 502, err.Error())
		return
	}
	if res.ResultCode == 7 {
		_ = ctx.Response(map[string]any{"confirm": res, "locale": locale})
		return
	}
	if res.ResultCode != 0 {
		landingErr(ctx, 502, landingVivaResultMsg(res))
		return
	}
	landingEmitNewOrder(v, phone, body.ProductCode, body.SmsScenario, locale)
	_ = ctx.Response(map[string]any{"confirm": res, "locale": locale})
}

func landingGetSubscriberInfoGETHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingSubscriberGET(v, ctx, "")
	}
}

func landingGetSubscriberInfoGETLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingSubscriberGET(v, ctx, ctx.Param("locale"))
	}
}

func runLandingSubscriberGET(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if _, err := utils.LandingPickLocale(pathLocale, ""); err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	if !ensureVivaPartner(v, ctx) {
		return
	}
	phone := normalizeMSISDN(ctx.Param("phoneNum"))
	if phone == "" {
		landingErr(ctx, 400, "phoneNum required")
		return
	}
	landingRespondSubscriberInfo(v, ctx, phone)
}

type landingSubscriberInfoBody struct {
	PhoneNum string `json:"phoneNum"`
	Locale   string `json:"locale,omitempty"`
}

func landingGetSubscriberInfoPOSTHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingSubscriberPOST(v, ctx, "")
	}
}

func landingGetSubscriberInfoPOSTLocalizedHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		runLandingSubscriberPOST(v, ctx, ctx.Param("locale"))
	}
}

func runLandingSubscriberPOST(v *viva_api.Viva, ctx types.HandlerContext, pathLocale string) {
	if !ensureVivaPartner(v, ctx) {
		return
	}
	var body landingSubscriberInfoBody
	ctx.Data(&body)
	if _, err := utils.LandingPickLocale(pathLocale, body.Locale); err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	phone, _, err := landingResolveMSISDN(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	landingRespondSubscriberInfo(v, ctx, phone)
}

func landingRespondSubscriberInfo(v *viva_api.Viva, ctx types.HandlerContext, phone string) {
	c, cancel := landingCtx()
	defer cancel()
	info, err := v.VivaPartner.GetSubscriberInfo(c, phone)
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

func ensureVivaPartner(v *viva_api.Viva, ctx types.HandlerContext) bool {
	if v.VivaPartner != nil {
		return true
	}
	landingErr(ctx, 503, "viva api not configured")
	return false
}

func landingProduct(v *viva_api.Viva, name string) (string, bool) {
	p := strings.TrimSpace(name)
	if p == "" {
		p = v.DefaultProductName
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

func landingVivaResultMsg(m *vivaclient.ResponseModel) string {
	if m == nil {
		return "viva: empty response body"
	}
	return fmt.Sprintf("viva subscription resultCode=%d message=%v", m.ResultCode, m.Message)
}

func normalizeMSISDN(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	return s
}

func landingResolveMSISDN(ctx types.HandlerContext, bodyPhone string) (string, bool, error) {
	var hdr map[string]string
	ctx.Headers(&hdr)
	for k, v := range hdr {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		for _, name := range landingMSISDNHeaderNames {
			if strings.EqualFold(strings.TrimSpace(k), name) {
				return normalizeMSISDN(v), true, nil
			}
		}
	}
	if p := normalizeMSISDN(bodyPhone); p != "" {
		return p, false, nil
	}
	return "", false, fmt.Errorf("phoneNum missing: set JSON phoneNum (manual entry) or MSISDN header e.g. X-MSISDN (Viva header enrichment)")
}

func landingEmitNewOrder(v *viva_api.Viva, phone, bodyProductCode, smsScenario, locale string) {
	pc := strings.TrimSpace(bodyProductCode)
	if pc == "" {
		pc = strings.TrimSpace(v.OrderProductCode)
	}
	service.SendOrderCreate(v, types.OrderTypeNew, phone, pc, smsScenario, locale)
}
