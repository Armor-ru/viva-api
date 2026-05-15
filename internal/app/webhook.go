package viva_api

import (
	"github.com/Armor-ru/sds-go/pkg/types"
)

func (s *Viva) extActivationRequest(ctx types.HandlerContext) {
	s.handleExtCreate(ctx, types.OrderTypeNew)
}

func (s *Viva) extActivation(ctx types.HandlerContext) {
	s.handleExtCreate(ctx, types.OrderTypeRenew)
}

func (s *Viva) extRemove(ctx types.HandlerContext) {
	s.handleExtCreate(ctx, types.OrderTypeCancel)
}

func (s *Viva) handleExtCreate(ctx types.HandlerContext, orderType types.OrderType) {
	data := ExtReq{}
	ctx.Data(&data)
	s.sendOrderCreate(orderType, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
	_ = ctx.Response("")
}
