package viva_api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
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
	}
	ctx.Data(&body)

	productName := strings.TrimSpace(body.ProductName)
	if productName == "" {
		landingErr(ctx, 400, "productName required")
		return
	}
	phone, err := landingPhone(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.InitSubscription(c, phone, productName)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva InitSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	_ = ctx.Response(map[string]interface{}{"init": res})
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
	}
	ctx.Data(&body)

	productName := strings.TrimSpace(body.ProductName)
	if productName == "" {
		landingErr(ctx, 400, "productName required")
		return
	}
	otp := strings.TrimSpace(body.OTP)
	if otp == "" {
		landingErr(ctx, 400, "otp required")
		return
	}
	productCode := strings.TrimSpace(body.ProductCode)
	if productCode == "" {
		landingErr(ctx, 400, "productCode required")
		return
	}
	phone, err := landingPhone(ctx, body.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.ConfirmSubscription(c, phone, productName, otp)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription")
		landingErr(ctx, 502, err.Error())
		return
	}
	if res.ResultCode != 0 {
		landingErr(ctx, 502, vivaSubErr(res))
		return
	}

	orderId, err := s.createOrder(types.OrderTypeNew, phone, productCode)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Str("productCode", productCode).Msg("create landing order")
		landingErr(ctx, 502, err.Error())
		return
	}

	_ = ctx.Response(map[string]interface{}{"confirm": res, "orderId": orderId})
}

func landingPhone(ctx types.HandlerContext, bodyPhone string) (string, error) {
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
	return "", fmt.Errorf("phoneNum missing: provide it in body or MSISDN header")
}

func landingErr(ctx types.HandlerContext, status int, msg string) {
	_ = ctx.Response(httpTransport.MsgResponse{
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
