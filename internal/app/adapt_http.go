package viva_api

import (
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
)

func (s *Viva) ExtAppPartnerProductActivationRequestHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeNew)
}

func (s *Viva) ExtAppPartnerProductActivationHandler(ctx types.HandlerContext) {
	data := ExtReq{}
	ctx.Data(&data)

	phone := normalizeAddr(data.PhoneNum)
	productCode := strings.TrimSpace(data.ProductCode)
	if phone == "" || productCode == "" {
		logger.Error().Msg("webhook activation: phoneNum and productCode are required")
		return
	}

	_, shortNumber, _ := s.catalog.ProductByExternalID(productCode)
	ctxPipeline := pipeline.Context{
		Phone:       phone,
		ProductCode: productCode,
		ShortNumber: shortNumber,
		Lang:        s.catalog.ResolveLang(productCode, ""),
	}

	if err := s.runScenario("webhook.activation", ctxPipeline); err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("webhook activation scenario failed")
		return
	}
	_ = ctx.Response("")
}

func (s *Viva) ExtAppPartnerProductRemoveHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeCancel)
}

func (s *Viva) handleCreate(ctx types.HandlerContext, orderType types.OrderType) {
	data := ExtReq{}
	ctx.Data(&data)

	if _, err := s.createOrder(orderType, data.PhoneNum, data.ProductCode, ""); err != nil {
		logger.Error().Msg("can not create order, " + err.Error())
		return
	}

	_ = ctx.Response("")
}
