package viva_api

import (
	"github.com/Armor-ru/sds-go/pkg/types"
)

func (s *Viva) ExtAppPartnerProductActivationRequestHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeNew)
}

func (s *Viva) ExtAppPartnerProductActivationHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeRenew)
}

func (s *Viva) ExtAppPartnerProductRemoveHandler(ctx types.HandlerContext) {
	s.handleCreate(ctx, types.OrderTypeCancel)
}

func (s *Viva) handleCreate(ctx types.HandlerContext, orderType types.OrderType) {
	data := ExtReq{}
	ctx.Data(&data)
	s.sendOrderCreate(orderType, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
	_ = ctx.Response("")
}
