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

const vivaRequestTimeout = 25 * time.Second

var msisdnHeaders = []string{"X-MSISDN", "X-Msisdn", "X-Phone-Number"}

func (s *Viva) LandingInitHandler(ctx types.HandlerContext) {
	if s.vivaClient == nil {
		landingErr(ctx, 503, "viva api not configured")
		return
	}

	var body struct {
		PhoneNum    string `json:"phoneNum"`
		ProductName string `json:"productName"`
		SkipConfirm bool   `json:"skipConfirm"`
		Locale      string `json:"locale,omitempty"`
	}
	ctx.Data(&body)

	locale, err := pickLocale(ctx.Param("locale"), body.Locale)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	if strings.TrimSpace(body.ProductName) == "" {
		landingErr(ctx, 400, "productName required")
		return
	}
	if !body.SkipConfirm {
		landingErr(ctx, 400, "skipConfirm must be true; confirm requires OTP")
		return
	}

	phone, err := pickPhone(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.InitSubscription(c, phone, strings.TrimSpace(body.ProductName))
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva InitSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	_ = ctx.Response(map[string]interface{}{"init": res, "locale": locale})
}

func (s *Viva) LandingConfirmHandler(ctx types.HandlerContext) {
	if s.vivaClient == nil {
		landingErr(ctx, 503, "viva api not configured")
		return
	}

	var body struct {
		PhoneNum    string `json:"phoneNum"`
		ProductName string `json:"productName"`
		OTP         string `json:"otp"`
		ProductCode string `json:"productCode"`
		SmsScenario string `json:"smsScenario,omitempty"`
		Locale      string `json:"locale,omitempty"`
	}
	ctx.Data(&body)

	locale, err := pickLocale(ctx.Param("locale"), body.Locale)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	if strings.TrimSpace(body.ProductName) == "" {
		landingErr(ctx, 400, "productName required")
		return
	}
	otp := strings.TrimSpace(body.OTP)
	if otp == "" {
		landingErr(ctx, 400, "otp required")
		return
	}

	phone, err := pickPhone(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.ConfirmSubscription(c, phone, strings.TrimSpace(body.ProductName), &otp)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	if res.ResultCode == 7 {
		_ = ctx.Response(map[string]interface{}{"confirm": res, "locale": locale})
		return
	}
	if res.ResultCode != 0 {
		landingErr(ctx, 502, vivaSubErr(res))
		return
	}
	if strings.TrimSpace(body.ProductCode) == "" {
		landingErr(ctx, 400, "productCode required")
		return
	}

	s.handleCreate(ctx, types.OrderTypeNew, phone, body.ProductCode, body.SmsScenario, locale)
	_ = ctx.Response(map[string]interface{}{"confirm": res, "locale": locale})
}

func pickPhone(ctx types.HandlerContext, bodyPhone string) (string, error) {
	var headers map[string]string
	ctx.Headers(&headers)
	for key, value := range headers {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		for _, name := range msisdnHeaders {
			if strings.EqualFold(strings.TrimSpace(key), name) {
				return strings.TrimPrefix(value, "+"), nil
			}
		}
	}
	if p := strings.TrimSpace(strings.TrimPrefix(bodyPhone, "+")); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("phoneNum missing: provide in body or via MSISDN header")
}

func pickLocale(pathLocale, bodyLocale string) (string, error) {
	if t := strings.TrimSpace(pathLocale); t != "" {
		return parseLocale(t)
	}
	return localeOrDefault(bodyLocale), nil
}

func parseLocale(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "en", "eng", "english":
		return "en", nil
	case "ru", "rus", "russian":
		return "ru", nil
	case "hy", "arm", "hye", "armenian":
		return "hy", nil
	default:
		return "", fmt.Errorf("unknown locale %q: use en, ru, hy", s)
	}
}

func localeOrDefault(s string) string {
	loc, err := parseLocale(s)
	if err != nil {
		return "hy"
	}
	return loc
}

func landingErr(ctx types.HandlerContext, status int, msg string) {
	_ = ctx.Response(httpt.MsgResponse{
		Status:  status,
		Payload: map[string]string{"error": msg},
	})
}

func vivaSubErr(res *vivaclient.ResponseModel) string {
	if res == nil {
		return "viva: empty response"
	}
	return fmt.Sprintf("viva subscription resultCode=%d message=%v", res.ResultCode, res.Message)
}
