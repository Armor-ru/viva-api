package service

import (
	"strings"
	"time"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
	"github.com/Armor-ru/viva-api/internal/app/utils"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/google/uuid"
)

func SendOrderCreate(v *viva_api.Viva, orderType types.OrderType, phone, externalID, smsScenario, smsLocale string) {
	if v.IntTransport == nil {
		logger.Warn().Msg("intTransport nil, skip order/create")
		return
	}
	phone = strings.TrimSpace(phone)
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		logger.Warn().Str("phone", phone).Str("orderType", string(orderType)).Msg("skip order/create: productCode empty")
		return
	}
	if phone == "" {
		logger.Warn().Msg("skip order/create: phone empty")
		return
	}

	orderId := uuid.NewSHA1(uuid.MustParse(v.AccountId), []byte(externalID+":"+phone)).String()

	var items []types.OrderItemRequest
	if orderType == types.OrderTypeNew {
		items = append(items, types.OrderItemRequest{
			Id:         uuid.NewString(),
			ExternalId: &externalID,
		})
	}

	wh := ""
	switch orderType {
	case types.OrderTypeNew:
		wh = viva_api.WHActivationReq
	case types.OrderTypeRenew:
		wh = viva_api.WHActivation
	case types.OrderTypeCancel:
		wh = viva_api.WHRemove
	}

	newOrder := types.OrderCreateRequest{
		Id:   orderId,
		Type: orderType,
		Fields: types.JSON{
			"phone": phone,
		},
		CustomData: types.JSON{
			viva_api.CDVivaWebhook: wh,
			viva_api.CDSmsScenario: strings.TrimSpace(smsScenario),
			viva_api.CDSmsLocale:   utils.LocaleOrDefault(smsLocale),
		},
		Items: items,
	}

	_, err := v.IntTransport.Send("order/create", newOrder, types.SendOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		logger.Error().Interface("payload", newOrder).Msg("can not create order, " + err.Error())
		return
	}
	logger.Info().Str("orderId", orderId).Str("type", string(orderType)).Str("phone", phone).Msg("order/create sent")
}
