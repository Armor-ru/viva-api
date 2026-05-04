package handlers

import (
	viva_api "github.com/Armor-ru/viva-api/internal/app"
	"github.com/Armor-ru/viva-api/internal/app/service"

	"github.com/Armor-ru/sds-go/pkg/types"
)

func extAppPartnerProductActivationRequestHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleExtCreate(v, ctx, types.OrderTypeNew)
	}
}

func extAppPartnerProductActivationHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleExtCreate(v, ctx, types.OrderTypeRenew)
	}
}

func extAppPartnerProductRemoveHandler(v *viva_api.Viva) func(types.HandlerContext) {
	return func(ctx types.HandlerContext) {
		handleExtCreate(v, ctx, types.OrderTypeCancel)
	}
}

func handleExtCreate(v *viva_api.Viva, ctx types.HandlerContext, orderType types.OrderType) {
	data := viva_api.ExtReq{}
	ctx.Data(&data)
	service.SendOrderCreate(v, orderType, data.PhoneNum, data.ProductCode, data.SmsScenario, data.Locale)
	_ = ctx.Response("")
}
