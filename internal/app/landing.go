package viva_api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	httpTransport "git.dev.armlab.pro/armor/sds-go/pkg/transport/http"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cast"
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
	logLandingInitReceived(req.PhoneNum, req.ProductName)

	if err := validate.Struct(req); err != nil {
		landingErr(ctx, http.StatusBadRequest, errs.WrapWithFields(
			fmt.Errorf("invalid subscription init request"),
			map[string]interface{}{"err": err.Error()},
		), "init", req.PhoneNum, req.ProductName)
		return
	}

	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		landingErr(ctx, http.StatusBadRequest, errs.WrapWithFields(
			err,
			map[string]interface{}{"bodyPhone": req.PhoneNum},
		), "init", req.PhoneNum, req.ProductName)
		return
	}
	req.PhoneNum = phone

	res, err := s.landingInit(req)
	if err != nil {
		landingErr(ctx, http.StatusBadGateway, err, "init", req.PhoneNum, req.ProductName)
		return
	}

	_ = ctx.Response(map[string]interface{}{"init": res})
}

func (s *Viva) landingConfirmHandler(ctx types.HandlerContext) {
	var req SubscriptionConfirm
	ctx.Data(&req)
	logLandingConfirmReceived(
		req.PhoneNum,
		req.ProductName,
		strings.TrimSpace(req.OTP) != "",
		strings.TrimSpace(req.ProductCode) != "",
	)

	if err := validate.Struct(req); err != nil {
		landingErr(ctx, http.StatusBadRequest, errs.WrapWithFields(
			fmt.Errorf("invalid subscription confirm request"),
			map[string]interface{}{"err": err.Error()},
		), "confirm", req.PhoneNum, req.ProductName)
		return
	}

	phone, err := landingPhone(ctx, req.PhoneNum)
	if err != nil {
		landingErr(ctx, http.StatusBadRequest, errs.WrapWithFields(
			err,
			map[string]interface{}{"bodyPhone": req.PhoneNum},
		), "confirm", req.PhoneNum, req.ProductName)
		return
	}
	req.PhoneNum = phone

	res, orderID, err := s.landingConfirm(req)
	if err != nil {
		landingErr(ctx, http.StatusBadGateway, err, "confirm", req.PhoneNum, req.ProductName)
		return
	}

	logLandingConfirmDone(req.PhoneNum, orderID)
	_ = ctx.Response(map[string]interface{}{"confirm": res, "orderId": orderID})
}

func (s *Viva) landingInit(req SubscriptionInitRequest) (*vivaclient.ResponseModel, error) {
	if s.vivaClient == nil {
		return nil, fmt.Errorf("viva api not configured")
	}

	logger.Info().
		Str("phone", req.PhoneNum).
		Str("product", req.ProductName).
		Msg("viva init subscription request")

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.InitSubscription(c, req.PhoneNum, req.ProductName)
	if err != nil {
		return nil, errs.WrapWithFields(
			err,
			map[string]interface{}{
				"step":    "init",
				"phone":   req.PhoneNum,
				"product": req.ProductName,
				"path":    "/api/Subscription/InitSubscription",
			},
		)
	}
	if res.ResultCode != 0 {
		return nil, errs.WrapWithFields(
			vivaLandingErr(res.Message),
			map[string]interface{}{
				"step":       "init",
				"phone":      req.PhoneNum,
				"product":    req.ProductName,
				"resultCode": res.ResultCode,
				"path":       "/api/Subscription/InitSubscription",
			},
		)
	}

	logger.Info().
		Str("phone", req.PhoneNum).
		Str("product", req.ProductName).
		Int("resultCode", res.ResultCode).
		Msg("viva init subscription success")
	return res, nil
}

func (s *Viva) landingConfirm(req SubscriptionConfirm) (*vivaclient.ResponseModel, string, error) {
	if s.vivaClient == nil {
		return nil, "", fmt.Errorf("viva api not configured")
	}

	logger.Info().
		Str("phone", req.PhoneNum).
		Str("productName", req.ProductName).
		Str("productCode", req.ProductCode).
		Msg("viva confirm subscription request")

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	res, err := s.vivaClient.ConfirmSubscription(c, req.PhoneNum, req.ProductName, req.OTP)
	if err != nil {
		return nil, "", errs.WrapWithFields(
			err,
			map[string]interface{}{
				"step":        "confirm",
				"phone":       req.PhoneNum,
				"product":     req.ProductName,
				"productName": req.ProductName,
				"productCode": req.ProductCode,
				"path":        "/api/Subscription/ConfirmSubscription",
			},
		)
	}
	if res.ResultCode != 0 {
		return nil, "", errs.WrapWithFields(
			vivaLandingErr(res.Message),
			map[string]interface{}{
				"step":        "confirm",
				"phone":       req.PhoneNum,
				"product":     req.ProductName,
				"productName": req.ProductName,
				"productCode": req.ProductCode,
				"resultCode":  res.ResultCode,
				"path":        "/api/Subscription/ConfirmSubscription",
			},
		)
	}

	logger.Info().
		Str("phone", req.PhoneNum).
		Str("productName", req.ProductName).
		Int("resultCode", res.ResultCode).
		Msg("viva confirm subscription success")

	orderID := s.getOrderId(req.PhoneNum, req.ProductCode)
	var orderType types.OrderType
	order, err := s.getOrder(orderID)
	switch {
	case err == nil && strings.TrimSpace(order.ID) != "":
		if s.isActiveOrder(order) && !cast.ToBool(order.CustomData["expired"]) {
			logger.Info().
				Str("phone", req.PhoneNum).
				Str("productCode", req.ProductCode).
				Str("orderId", orderID).
				Str("orderAction", "reuse").
				Msg("landing active order reused")
			return res, orderID, nil
		}
		orderType = types.OrderTypeRenew
	case err != nil && !strings.Contains(strings.ToLower(err.Error()), "order not found"):
		return nil, "", errs.WrapWithFields(
			err,
			map[string]interface{}{
				"step":        "confirm",
				"phone":       req.PhoneNum,
				"product":     req.ProductCode,
				"productName": req.ProductName,
				"productCode": req.ProductCode,
				"orderId":     orderID,
				"path":        "order/get",
			},
		)
	default:
		orderType = types.OrderTypeNew
	}

	logger.Info().
		Str("phone", req.PhoneNum).
		Str("productCode", req.ProductCode).
		Str("orderId", orderID).
		Str("orderType", string(orderType)).
		Str("orderAction", "create").
		Msg("landing order type resolved")

	lang := normalizeLangCode(req.Lang)
	if lang == "" {
		lang = s.GetLang(req.PhoneNum)
	}
	if lang == "" {
		lang = defaultLang
	}

	logger.Info().
		Str("phone", req.PhoneNum).
		Str("productCode", req.ProductCode).
		Str("orderId", orderID).
		Str("orderType", string(orderType)).
		Str("lang", lang).
		Msg("landing order create sent")
	if err := s.createOrder(orderType, req.PhoneNum, req.ProductCode, lang); err != nil {
		return nil, "", errs.WrapWithFields(
			err,
			map[string]interface{}{
				"step":        "confirm",
				"phone":       req.PhoneNum,
				"product":     req.ProductCode,
				"productName": req.ProductName,
				"productCode": req.ProductCode,
				"orderId":     orderID,
				"orderType":   orderType,
				"path":        "order/create",
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
	existing := errs.Fields(err)
	wrap := map[string]interface{}{}
	if existing == nil || existing["step"] == nil {
		wrap["step"] = step
	}
	if existing == nil || existing["phone"] == nil {
		wrap["phone"] = phone
	}
	if existing == nil || (existing["product"] == nil && existing["productName"] == nil && existing["productCode"] == nil) {
		wrap["productName"] = product
	}
	if len(wrap) > 0 {
		err = errs.WrapWithFields(err, wrap)
	}
	payload := landingErrorPayload(err)
	status = landingHTTPStatus(status, err)
	logAppError(err, landingErrMsg(step))
	_ = ctx.Response(httpTransport.MsgResponse{
		Status:  status,
		Payload: payload,
	})
}

func landingErrorPayload(err error) LandingErrorResponse {
	payload := LandingErrorResponse{Error: err.Error()}
	if resultCode, ok := landingResultCode(err); ok {
		payload.ResultCode = &resultCode
	}
	return payload
}

func landingHTTPStatus(fallback int, err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}

	resultCode, ok := landingResultCode(err)
	if !ok {
		return fallback
	}

	switch resultCode {
	case 1:
		return http.StatusForbidden
	case 2, 9, 20:
		return http.StatusNotFound
	case 7, 17, 18:
		return http.StatusConflict
	case 12, 16:
		return http.StatusBadRequest
	case 23:
		return http.StatusTooManyRequests
	case 6, 10, 11, 13, 21:
		return http.StatusBadGateway
	case 3, 4, 5, 8, 14, 15, 19, 22:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusUnprocessableEntity
	}
}

func landingResultCode(err error) (int, bool) {
	fields := errs.Fields(err)
	if fields == nil {
		return 0, false
	}
	resultCode, ok := fields["resultCode"].(int)
	return resultCode, ok
}

func landingErrMsg(step string) string {
	switch step {
	case "init":
		return "viva init subscription from landing failed"
	case "confirm":
		return "viva confirm subscription from landing failed"
	default:
		return "landing " + step + " failed"
	}
}

func vivaLandingErr(message interface{}) error {
	msg := strings.TrimSpace(strings.ToLower(fmt.Sprint(message)))
	if msg == "" {
		return errors.New("viva subscription rejected")
	}
	return errors.New(msg)
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
