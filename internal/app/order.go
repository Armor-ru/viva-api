package viva_api

import (
	"strings"
	"time"

	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/types"
	"github.com/google/uuid"
)

func (s *Viva) sendOrderCreate(orderType types.OrderType, phone, externalID, smsScenario, smsLocale string) {
	if s.intTransport == nil {
		logger.Warn().Msg("intTransport nil, skip order/create")
		return
	}
	phone = strings.TrimSpace(phone)
	externalID = strings.TrimSpace(externalID)
	if externalID == "" || phone == "" {
		logger.Warn().Str("phone", phone).Str("orderType", string(orderType)).Msg("skip order/create: empty phone or productCode")
		return
	}

	orderId := uuid.NewSHA1(uuid.MustParse(s.accountId), []byte(externalID+":"+phone)).String()

	var items []types.OrderItemRequest
	if orderType == types.OrderTypeNew {
		ext := externalID
		items = append(items, types.OrderItemRequest{
			Id:         uuid.NewString(),
			ExternalId: &ext,
		})
	}

	wh := ""
	switch orderType {
	case types.OrderTypeNew:
		wh = WHActivationReq
	case types.OrderTypeRenew:
		wh = WHActivation
	case types.OrderTypeCancel:
		wh = WHRemove
	}

	newOrder := types.OrderCreateRequest{
		Id:   orderId,
		Type: orderType,
		Fields: types.JSON{
			"phone": phone,
		},
		CustomData: types.JSON{
			CDVivaWebhook: wh,
			CDSmsScenario: strings.TrimSpace(smsScenario),
			CDSmsLocale:   localeOrDefault(smsLocale),
		},
		Items: items,
	}

	_, err := s.intTransport.Send("order/create", newOrder, types.SendOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		logger.Error().Interface("payload", newOrder).Msg("can not create order, " + err.Error())
		return
	}
	logger.Info().Str("orderId", orderId).Str("type", string(orderType)).Str("phone", phone).Msg("order/create sent")
}
