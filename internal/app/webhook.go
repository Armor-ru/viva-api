package viva_api

import (
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func (s *viva) webhookHandler(ctx types.HandlerContext, orderType types.OrderType) {
	data := ExtReq{}
	ctx.Data(&data)

	if err := s.createOrder(orderType, data.PhoneNum, data.ProductCode, ""); err != nil {
		logger.Error().Err(err).Msg("webhook create order failed")
		return
	}

	_ = ctx.Response("")
}
