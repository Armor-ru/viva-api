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

type SubscriptionInitRequest struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName"`
	Lang        string `json:"lang"`
}

type SubscriptionConfirmRequest struct {
	PhoneNum    string `json:"phoneNum"`
	ProductName string `json:"productName"`
	OTP         string `json:"otp"`
	ProductCode string `json:"productCode"`
	Lang        string `json:"lang"`
}

func (s *viva) landingInitHandler(ctx types.HandlerContext) {
	var req SubscriptionInitRequest
	ctx.Data(&req)
	res, err := s.landingInit(req, ctx)
	if err != nil {
		landingErr(ctx, landingStatus(err), err.Error())
		return
	}
	_ = ctx.Response(map[string]interface{}{"init": res})
}

func (s *viva) landingConfirmHandler(ctx types.HandlerContext) {
	var req SubscriptionConfirmRequest
	ctx.Data(&req)
	res, orderID, err := s.landingConfirm(req, ctx)
	if err != nil {
		landingErr(ctx, landingStatus(err), err.Error())
		return
	}
	_ = ctx.Response(map[string]interface{}{"confirm": res, "orderId": orderID})
}

func (s *viva) landingInit(req SubscriptionInitRequest, ctx types.HandlerContext) (interface{}, error) {
	if s.vivaClient == nil {
		return nil, landingHTTPError{status: 503, msg: "viva api not configured"}
	}
	productName := strings.TrimSpace(req.ProductName)
	if productName == "" {
		return nil, landingHTTPError{status: 400, msg: "productName required"}
	}
	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		return nil, landingHTTPError{status: 400, msg: err.Error()}
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.InitSubscription(c, phone, productName)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva InitSubscription")
		return nil, landingHTTPError{status: 502, msg: err.Error()}
	}
	return res, nil
}

func (s *viva) landingConfirm(req SubscriptionConfirmRequest, ctx types.HandlerContext) (interface{}, string, error) {
	if s.vivaClient == nil {
		return nil, "", landingHTTPError{status: 503, msg: "viva api not configured"}
	}
	productName := strings.TrimSpace(req.ProductName)
	if productName == "" {
		return nil, "", landingHTTPError{status: 400, msg: "productName required"}
	}
	otp := strings.TrimSpace(req.OTP)
	if otp == "" {
		return nil, "", landingHTTPError{status: 400, msg: "otp required"}
	}
	productCode := strings.TrimSpace(req.ProductCode)
	if productCode == "" {
		return nil, "", landingHTTPError{status: 400, msg: "productCode required"}
	}
	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		return nil, "", landingHTTPError{status: 400, msg: err.Error()}
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.ConfirmSubscription(c, phone, productName, otp)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Viva ConfirmSubscription")
		return nil, "", landingHTTPError{status: 502, msg: err.Error()}
	}
	if res.ResultCode != 0 {
		return nil, "", landingHTTPError{status: 502, msg: vivaSubErr(res)}
	}

	lang := strings.TrimSpace(req.Lang)
	if lang == "" {
		lang = s.resolveLang(phone, s.defaultLanguageForProductCode(productCode))
	}

	orderID := s.getOrderId(phone, productCode)
	if err := s.createOrder(types.OrderTypeNew, phone, productCode, lang); err != nil {
		logger.Error().Err(err).Str("phone", phone).Str("productCode", productCode).Msg("create landing order")
		return nil, "", landingHTTPError{status: 502, msg: err.Error()}
	}
	return res, orderID, nil
}

type landingHTTPError struct {
	status int
	msg    string
}

func (e landingHTTPError) Error() string { return e.msg }

func landingStatus(err error) int {
	if he, ok := err.(landingHTTPError); ok {
		return he.status
	}
	return 502
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
				return normalizePhone(value), nil
			}
		}
	}
	if p := normalizePhone(bodyPhone); p != "" {
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
