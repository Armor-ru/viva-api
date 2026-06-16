package viva_api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
	"github.com/go-playground/validator/v10"
)

const vivaRequestTimeout = 25 * time.Second

var validate = validator.New()

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

	if err := validate.Struct(req); err != nil {
		landingErr(ctx, 400, errs.WrapWithFields(
			fmt.Errorf("invalid subscription init request"),
			map[string]interface{}{"err": err.Error()},
		), "init", req.PhoneNum, req.ProductName)
		return
	}

	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, errs.WrapWithFields(
			err,
			map[string]interface{}{"bodyPhone": req.PhoneNum},
		), "init", req.PhoneNum, req.ProductName)
		return
	}
	req.PhoneNum = phone

	res, err := s.landingInit(req)
	if err != nil {
		landingErr(ctx, 502, err, "init", req.PhoneNum, req.ProductName)
		return
	}

	_ = ctx.Response(map[string]interface{}{"init": res})
}

func (s *Viva) landingConfirmHandler(ctx types.HandlerContext) {
	var req SubscriptionConfirm
	ctx.Data(&req)

	if err := validate.Struct(req); err != nil {
		landingErr(ctx, 400, errs.WrapWithFields(
			fmt.Errorf("invalid subscription confirm request"),
			map[string]interface{}{"err": err.Error()},
		), "confirm", req.PhoneNum, req.ProductName)
		return
	}

	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		landingErr(ctx, 400, errs.WrapWithFields(
			err,
			map[string]interface{}{"bodyPhone": req.PhoneNum},
		), "confirm", req.PhoneNum, req.ProductName)
		return
	}
	req.PhoneNum = phone

	res, orderID, err := s.landingConfirm(req)
	if err != nil {
		landingErr(ctx, 502, err, "confirm", req.PhoneNum, req.ProductName)
		return
	}

	logLandingConfirmDone(req.PhoneNum, orderID)
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
		return nil, errs.WrapWithFields(
			err,
			map[string]interface{}{
				"phone":   req.PhoneNum,
				"product": req.ProductName,
			},
		)
	}
	if res.ResultCode != 0 {
		return nil, errs.WrapWithFields(
			fmt.Errorf("viva init subscription failed"),
			map[string]interface{}{
				"phone":      req.PhoneNum,
				"product":    req.ProductName,
				"resultCode": res.ResultCode,
				"message":    res.Message,
			},
		)
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
		return nil, "", errs.WrapWithFields(
			err,
			map[string]interface{}{
				"phone":   req.PhoneNum,
				"product": req.ProductName,
			},
		)
	}
	if res.ResultCode != 0 {
		return nil, "", errs.WrapWithFields(
			fmt.Errorf("viva subscription failed"),
			map[string]interface{}{
				"phone":      req.PhoneNum,
				"product":    req.ProductName,
				"resultCode": res.ResultCode,
				"message":    res.Message,
			},
		)
	}

	orderID := s.getOrderId(req.PhoneNum, req.ProductCode)
	if err := s.createOrder(types.OrderTypeNew, req.PhoneNum, req.ProductCode, req.Lang); err != nil {
		return nil, "", errs.WrapWithFields(
			err,
			map[string]interface{}{
				"phone":   req.PhoneNum,
				"product": req.ProductCode,
			},
		)
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
	return "", errs.WrapWithFields(
		fmt.Errorf("phoneNum missing"),
		map[string]interface{}{"bodyPhone": bodyPhone},
	)
}

func landingErr(ctx types.HandlerContext, status int, err error, step, phone, product string) {
	fields := map[string]interface{}{
		"step":    step,
		"phone":   phone,
		"product": product,
	}
	logAppError(errs.WrapWithFields(err, fields), "landing "+step+" failed")
	_ = ctx.Response(httpTransport.MsgResponse{
		Status:  status,
		Payload: map[string]string{"error": err.Error()},
	})
}

func (s *Viva) removeSubscription(phoneNum, productCode string) error {
	if s.vivaClient == nil {
		return fmt.Errorf("viva api not configured")
	}
	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()
	res, err := s.vivaClient.RemoveSubscription(c, phoneNum, productCode)
	if err != nil {
		return errs.WrapWithFields(
			err,
			map[string]interface{}{
				"phone":   phoneNum,
				"product": productCode,
			},
		)
	}
	if res != nil && res.ResultCode != 0 {
		return errs.WrapWithFields(
			fmt.Errorf("viva remove subscription rejected"),
			map[string]interface{}{
				"phone":      phoneNum,
				"product":    productCode,
				"resultCode": res.ResultCode,
				"message":    res.Message,
			},
		)
	}
	return nil
}
