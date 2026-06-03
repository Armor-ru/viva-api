package viva_api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"github.com/google/uuid"
	"github.com/spf13/cast"
)

func (s *viva) getOrderId(phone, productCode string) string {
	phone = strings.TrimSpace(phone)
	productCode = strings.TrimSpace(productCode)
	return uuid.NewSHA1(uuid.MustParse(s.accountId), []byte(productCode+":"+phone)).String()
}

func (s *viva) getOrder(id string) (Order, error) {
	if s.intTransport == nil {
		return Order{}, fmt.Errorf("intTransport is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return Order{}, fmt.Errorf("order id is required")
	}

	raw, err := s.intTransport.Send("order/get", types.GetOrderRequest{ID: id}, types.SendOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return Order{}, fmt.Errorf("send order/get: %w", err)
	}

	var order Order
	switch v := raw.(type) {
	case Order:
		return v, nil
	case map[string]interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return Order{}, err
		}
		if err := json.Unmarshal(b, &order); err != nil {
			return Order{}, err
		}
		return order, nil
	default:
		b, err := json.Marshal(raw)
		if err != nil {
			return Order{}, err
		}
		if err := json.Unmarshal(b, &order); err != nil {
			return Order{}, fmt.Errorf("decode order/get response: %w", err)
		}
		return order, nil
	}
}

type orderLookup struct {
	Exists bool
	Active bool
	Order  Order
}

func (s *viva) lookupOrder(orderID string) orderLookup {
	order, err := s.getOrder(orderID)
	if err != nil {
		logger.Info().Err(err).Str("orderId", orderID).Msg("order/get: order not found")
		return orderLookup{Exists: false}
	}
	if strings.TrimSpace(order.ID) == "" {
		return orderLookup{Exists: false}
	}
	active := s.isActiveOrder(order)
	logger.Info().
		Str("orderId", order.ID).
		Str("status", order.Status).
		Bool("active", active).
		Msg("order/get: order exists")
	return orderLookup{Exists: true, Active: active, Order: order}
}

func (s *viva) isActiveOrder(order Order) bool {
	if strings.TrimSpace(order.Status) != string(types.OrderStatusListCompleted) {
		return false
	}
	if order.EndTime == nil {
		return false
	}
	return order.EndTime.After(time.Now())
}

func (s *viva) createOrder(orderType types.OrderType, phoneNum, productCode, lang string) error {
	return s.createOrderWithID(orderType, "", phoneNum, productCode, lang)
}

func (s *viva) createOrderWithID(orderType types.OrderType, orderID, phoneNum, productCode, lang string) error {
	if s.intTransport == nil {
		return fmt.Errorf("intTransport is not configured")
	}

	phoneNum = strings.TrimSpace(phoneNum)
	productCode = strings.TrimSpace(productCode)
	if phoneNum == "" || productCode == "" {
		return fmt.Errorf("phoneNum and productCode are required")
	}

	if strings.TrimSpace(orderID) == "" {
		orderID = s.getOrderId(phoneNum, productCode)
	}

	lang = s.resolveLang(phoneNum, lang)
	req := buildOrderCreateRequest(orderType, orderID, phoneNum, productCode, lang)
	return s.publishOrderCreate(req)
}

func buildOrderCreateRequest(orderType types.OrderType, orderID, phoneNum, productCode, lang string) types.OrderCreateRequest {
	var items []types.OrderItemRequest
	if orderType == types.OrderTypeNew {
		items = append(items, types.OrderItemRequest{
			ID:         uuid.NewString(),
			ExternalID: &productCode,
		})
	}

	customData := types.JSON{"productCode": productCode}
	if lang != "" {
		customData["lang"] = lang
	}

	return types.OrderCreateRequest{
		ID:         orderID,
		Type:       orderType,
		Fields:     types.JSON{"phone": phoneNum},
		CustomData: customData,
		Items:      items,
	}
}

func (s *viva) publishOrderCreate(req types.OrderCreateRequest) error {
	_, err := s.intTransport.Send("order/create", req, types.SendOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		logger.Error().Interface("payload", req).Msg("can not create order, " + err.Error())
		return fmt.Errorf("send order/create: %w", err)
	}
	logger.Info().
		Str("orderId", req.ID).
		Str("type", string(req.Type)).
		Str("topic", "order/create").
		Msg("order/create published to NATS")
	return nil
}

func (s *viva) completeOrder(order types.OrderResponse) error {
	if order.Status == "error" {
		logger.Info().Str("orderId", order.ID).Msg("order status error, no need to send notify")
		return nil
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("can not send notify, order items is empty")
	}

	item, ok := firstActivateOrderItem(order.Items)
	if !ok {
		logger.Info().Str("orderId", order.ID).Msg("no need to send notify with activation code")
		return nil
	}

	artifacts, err := activationArtifactsFromItem(item)
	if err != nil {
		return err
	}
	activationCode := artifacts.ActivationCode
	downloadURL := artifacts.DownloadURL

	logger.Info().
		Str("orderId", order.ID).
		Str("activationCode", activationCode).
		Str("downloadURL", downloadURL).
		Msg("scenario step 11: activation code and download link received from order/completed")

	rawPhone, ok := order.Fields["phone"]
	if !ok {
		return fmt.Errorf("can not send notify, order has no phone")
	}
	phone := strings.TrimSpace(cast.ToString(rawPhone))
	if phone == "" {
		return fmt.Errorf("can not send notify, phone is empty")
	}

	lang := ""
	if order.CustomData != nil {
		lang = cast.ToString(order.CustomData["lang"])
	}
	lang = s.resolveLang(phone, lang)

	productCode := productCodeFromOrder(order, item)
	product, err := s.productFromCatalogByExternalID(productCode)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Phone":          phone,
		"ExternalID":     productCode,
		"ActivationCode": activationCode,
		"DownloadURL":    downloadURL,
		"Link":           downloadURL,
		"Language":       lang,
	}

	return s.sendActivationSMS(phone, lang, product, data)
}

func firstDownloadURL(raw interface{}) (string, error) {
	list, ok := raw.([]interface{})
	if !ok {
		list = []interface{}{}
	}
	for _, v := range list {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if url := strings.TrimSpace(cast.ToString(m["url"])); url != "" {
			return url, nil
		}
	}
	return "", fmt.Errorf("can not send notify, artifacts download not found")
}
