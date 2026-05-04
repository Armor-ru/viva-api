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
	ProductCode string `json:"productCode"`
	SmsScenario string `json:"smsScenario,omitempty"`
	// Locale — en | ru | hy (или eng / rus / arm); для SMS. Если пусто — ru.
	Locale string `json:"locale,omitempty"`
}

type LandingConfirmBody struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName"`
	OTP         string `json:"otp" validate:"required"`
	ProductCode string `json:"productCode"`
	SmsScenario string `json:"smsScenario,omitempty"`
	Locale      string `json:"locale,omitempty"`
}

func (s *Viva) LandingInitSubscriptionHandler(ctx types.HandlerContext) {
	s.runLandingInit(ctx, "")
}

func (s *Viva) LandingInitSubscriptionLocalizedHandler(ctx types.HandlerContext) {
	s.runLandingInit(ctx, ctx.Param("locale"))
}

func (s *Viva) runLandingInit(ctx types.HandlerContext, pathLocale string) {
	if !s.ensureVivaPartner(ctx) {
		return
	}
	var body LandingInitBody
	ctx.Data(&body)

	locale, err := landingPickLocale(pathLocale, body.Locale)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	prod, ok := s.landingProduct(body.ProductName)
	if !ok {
		landingErr(ctx, 400, "productName required")
		return
	}
	phone, msisdnFromHeader, err := landingResolveMSISDN(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	// Без доверенного MSISDN в заголовке (обогащение у Viva) номер из JSON — это ручной ввод:
	// серверный Confirm без OTP запрещён, иначе любой мог бы оформить подписку на чужой номер.
	if !msisdnFromHeader && !body.SkipConfirm {
		landingErr(ctx, 400, "without carrier MSISDN header (e.g. X-MSISDN), skipConfirm must be true; complete flow with POST .../confirm-subscription and otp")
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
		_ = ctx.Response(map[string]any{"init": initRes, "locale": locale})
		return
	}

	confirmRes, err := s.vivaPartner.ConfirmSubscription(c, phone, prod, nil)
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
	s.landingEmitNewOrder(phone, body.ProductCode, body.SmsScenario, locale)
	_ = ctx.Response(map[string]any{"init": initRes, "confirm": confirmRes, "locale": locale})
}

func (s *Viva) LandingConfirmSubscriptionHandler(ctx types.HandlerContext) {
	s.runLandingConfirm(ctx, "")
}

func (s *Viva) LandingConfirmSubscriptionLocalizedHandler(ctx types.HandlerContext) {
	s.runLandingConfirm(ctx, ctx.Param("locale"))
}

func (s *Viva) runLandingConfirm(ctx types.HandlerContext, pathLocale string) {
	if !s.ensureVivaPartner(ctx) {
		return
	}
	var body LandingConfirmBody
	ctx.Data(&body)

	locale, err := landingPickLocale(pathLocale, body.Locale)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

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
	phone, _, err := landingResolveMSISDN(ctx, body.PhoneNum)
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
	if res.ResultCode == 7 {
		_ = ctx.Response(map[string]any{"confirm": res, "locale": locale})
		return
	}
	if res.ResultCode != 0 {
		landingErr(ctx, 502, landingVivaResultMsg(res))
		return
	}
	s.landingEmitNewOrder(phone, body.ProductCode, body.SmsScenario, locale)
	_ = ctx.Response(map[string]any{"confirm": res, "locale": locale})
}

func (s *Viva) LandingGetSubscriberInfoGETHandler(ctx types.HandlerContext) {
	s.runLandingSubscriberGET(ctx, "")
}

func (s *Viva) LandingGetSubscriberInfoGETLocalizedHandler(ctx types.HandlerContext) {
	s.runLandingSubscriberGET(ctx, ctx.Param("locale"))
}

func (s *Viva) runLandingSubscriberGET(ctx types.HandlerContext, pathLocale string) {
	if _, err := landingPickLocale(pathLocale, ""); err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
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
	Locale   string `json:"locale,omitempty"`
}

func (s *Viva) LandingGetSubscriberInfoPOSTHandler(ctx types.HandlerContext) {
	s.runLandingSubscriberPOST(ctx, "")
}

func (s *Viva) LandingGetSubscriberInfoPOSTLocalizedHandler(ctx types.HandlerContext) {
	s.runLandingSubscriberPOST(ctx, ctx.Param("locale"))
}

func (s *Viva) runLandingSubscriberPOST(ctx types.HandlerContext, pathLocale string) {
	if !s.ensureVivaPartner(ctx) {
		return
	}
	var body landingSubscriberInfoBody
	ctx.Data(&body)
	if _, err := landingPickLocale(pathLocale, body.Locale); err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	phone, _, err := landingResolveMSISDN(ctx, body.PhoneNum)
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

// landingResolveMSISDN возвращает номер и признак: msisdnFromHeader=true, если номер взят из заголовка обогащения (вариант A).
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

func (s *Viva) landingEmitNewOrder(phone, bodyProductCode, smsScenario, locale string) {
	pc := strings.TrimSpace(bodyProductCode)
	if pc == "" {
		pc = strings.TrimSpace(s.orderProductCode)
	}
	s.sendOrderCreate(types.OrderTypeNew, phone, pc, smsScenario, locale)
}
