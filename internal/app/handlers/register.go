package handlers

import (
	viva_api "github.com/Armor-ru/viva-api/internal/app"
	"github.com/Armor-ru/viva-api/internal/app/middleware"
	"github.com/Armor-ru/viva-api/internal/app/service"

	"github.com/Armor-ru/sds-go/pkg/types"
)

func Register(v *viva_api.Viva) {
	middleware.InitExtHTTP(v)

	if v.ExtTransport != nil {
		v.ExtTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", extAppPartnerProductActivationRequestHandler(v))
		v.ExtTransport.Subscribe("POST /ExtAppPartneerProductActivation", extAppPartnerProductActivationHandler(v))
		v.ExtTransport.Subscribe("POST /ExtAppPartneerProductRemove", extAppPartnerProductRemoveHandler(v))

		v.ExtTransport.Subscribe("POST /landing/init-subscription", landingInitSubscriptionHandler(v))
		v.ExtTransport.Subscribe("POST /landing/confirm-subscription", landingConfirmSubscriptionHandler(v))
		v.ExtTransport.Subscribe("GET /landing/subscriber-info/:phoneNum", landingGetSubscriberInfoGETHandler(v))
		v.ExtTransport.Subscribe("POST /landing/subscriber-info", landingGetSubscriberInfoPOSTHandler(v))

		v.ExtTransport.Subscribe("POST /landing/:locale/init-subscription", landingInitSubscriptionLocalizedHandler(v))
		v.ExtTransport.Subscribe("POST /landing/:locale/confirm-subscription", landingConfirmSubscriptionLocalizedHandler(v))
		v.ExtTransport.Subscribe("GET /landing/:locale/subscriber-info/:phoneNum", landingGetSubscriberInfoGETLocalizedHandler(v))
		v.ExtTransport.Subscribe("POST /landing/:locale/subscriber-info", landingGetSubscriberInfoPOSTLocalizedHandler(v))
	}

	if v.IntTransport != nil {
		v.IntTransport.Subscribe("order/completed", func(ctx types.HandlerContext) {
			service.HandleOrderCompleted(v, ctx)
		})
	}
}
