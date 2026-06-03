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

type SubscriptionInitRequest struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName" validate:"required"`
	Lang        string `json:"lang"`
}

type SubscriptionConfirm struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName" validate:"required"`
	OTP         string `json:"otp" validate:"required"`
	ProductCode string `json:"productCode" validate:"required"`
	Lang        string `json:"lang"`
}

func (s *Viva) landingInitHandler(ctx types.HandlerContext) {
	var req SubscriptionInitRequest
	ctx.Data(&req)

	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	req.PhoneNum = phone

	res, err := s.landingInit(req)
	if err != nil {
		landingErr(ctx, 502, err.Error())
		return
	}
	_ = ctx.Response(map[string]interface{}{"init": res})
}

func (s *Viva) landingConfirmHandler(ctx types.HandlerContext) {
	var req SubscriptionConfirm
	ctx.Data(&req)

	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, err.Error())
		return
	}
	req.PhoneNum = phone

	res, orderID, err := s.landingConfirm(req)
	if err != nil {
		landingErr(ctx, 502, err.Error())
		return
	}
	_ = ctx.Response(map[string]interface{}{"confirm": res, "orderId": orderID})
}

func (s *Viva) landingInit(req SubscriptionInitRequest) (*vivaclient.ResponseModel, error) {
	if s.vivaClient == nil {
		return nil, fmt.Errorf("viva api not configured")
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.InitSubscription(c, req.PhoneNum, req.ProductName)
	if err != nil {
		logger.Error().Err(err).Str("phone", req.PhoneNum).Msg("Viva InitSubscription")
		return nil, err
	}
	return res, nil
}

func (s *Viva) landingConfirm(req SubscriptionConfirm) (*vivaclient.ResponseModel, string, error) {
	if s.vivaClient == nil {
		return nil, "", fmt.Errorf("viva api not configured")
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.ConfirmSubscription(c, req.PhoneNum, req.ProductName, req.OTP)
	if err != nil {
		logger.Error().Err(err).Str("phone", req.PhoneNum).Msg("Viva ConfirmSubscription")
		return nil, "", err
	}
	if res.ResultCode != 0 {
		return nil, "", fmt.Errorf("viva subscription resultCode=%d message=%v", res.ResultCode, res.Message)
	}

	orderID := s.getOrderId(req.PhoneNum, req.ProductCode)
	if err := s.createOrder(types.OrderTypeNew, req.PhoneNum, req.ProductCode, req.Lang); err != nil {
		logger.Error().Err(err).Str("phone", req.PhoneNum).Str("productCode", req.ProductCode).Msg("create landing order")
		return nil, "", err
	}
	return res, orderID, nil
}

func landingPhone(ctx types.HandlerContext, bodyPhone string) (string, error) {
	var headers map[string]string
	ctx.Headers(&headers)
	for key, value := range headers {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "x-msisdn", "x-phone-number":
			return strings.TrimPrefix(value, "+"), nil
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
