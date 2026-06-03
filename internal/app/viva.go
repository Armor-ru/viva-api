package viva_api

import (
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
)

// Order — снимок заказа SDS (pkg/types).
type Order = types.OrderResponse

// Viva — контракт сервиса по SPEC тимлида.
type Viva interface {
	InitHandlers()
	ussdHandler(ctx types.HandlerContext)
	orderCompleteHandler(ctx types.HandlerContext)
	webhookHandler(ctx types.HandlerContext, orderType types.OrderType)
	landingInitHandler(ctx types.HandlerContext)
	landingConfirmHandler(ctx types.HandlerContext)
}

// viva — реализация Viva (docs/SPEC.md, review: docs/SCENARIOS.md).
type viva struct {
	intTransport      types.Transport
	extTransport      types.Transport
	ussdTransport     types.Transport
	vivaClient        *vivaclient.Client
	langStore         LangStore
	paidWelcomeSent   PaidWelcomeStore
	catalogDir        string
	catalog           catalog.Catalog
	secrets           []string
	accountId         string
	landingConfirmURL string
}

var _ Viva = (*viva)(nil)
