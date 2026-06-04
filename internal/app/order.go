package viva_api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"github.com/google/uuid"
	"github.com/spf13/cast"
)

type Order = types.OrderResponse

func (s *Viva) getOrderId(phone, productCode string) string {
	return uuid.NewSHA1(
		uuid.MustParse(s.accountId),
		[]byte(strings.TrimSpace(productCode)+":"+strings.TrimSpace(phone)),
	).String()
}

func (s *Viva) getOrder(id string) (Order, error) {
	if s.intTransport == nil {
		return Order{}, errs.WrapWithFields(
			fmt.Errorf("intTransport is not configured"),
			map[string]interface{}{"orderId": id},
		)
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return Order{}, errs.WrapWithFields(
			fmt.Errorf("order id is required"),
			map[string]interface{}{"orderId": id},
		)
	}

	raw, err := s.intTransport.Send("order/get", types.GetOrderRequest{ID: id}, types.SendOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return Order{}, errs.WrapWithFields(
			fmt.Errorf("send order/get failed %w", err),
			map[string]interface{}{"orderId": id},
		)
	}

	var order Order
	data, err := json.Marshal(raw)
	if err != nil {
		return Order{}, errs.WrapWithFields(
			fmt.Errorf("marshal order/get response failed %w", err),
			map[string]interface{}{"orderId": id},
		)
	}
	if err := json.Unmarshal(data, &order); err != nil {
		return Order{}, errs.WrapWithFields(
			fmt.Errorf("unmarshal order/get response failed %w", err),
			map[string]interface{}{"orderId": id},
		)
	}
	return order, nil
}

func (s *Viva) isActiveOrder(order Order) bool {
	if strings.TrimSpace(order.Status) != string(types.OrderStatusListCompleted) {
		return false
	}
	if order.EndTime == nil {
		return false
	}
	return order.EndTime.After(time.Now())
}

func (s *Viva) createOrder(orderType types.OrderType, phoneNum, productCode, lang string) error {
	if s.intTransport == nil {
		return errs.WrapWithFields(
			fmt.Errorf("intTransport is not configured"),
			map[string]interface{}{"phoneNum": phoneNum, "productCode": productCode},
		)
	}

	phoneNum = strings.TrimSpace(phoneNum)
	productCode = strings.TrimSpace(productCode)
	if phoneNum == "" || productCode == "" {
		return errs.WrapWithFields(
			fmt.Errorf("phoneNum and productCode are required"),
			map[string]interface{}{"phoneNum": phoneNum, "productCode": productCode},
		)
	}

	lang = strings.TrimSpace(lang)
	if lang == "" {
		lang = s.storedLang(phoneNum)
	}

	req := types.OrderCreateRequest{
		ID:     s.getOrderId(phoneNum, productCode),
		Type:   orderType,
		Fields: types.JSON{"phone": phoneNum},
		CustomData: types.JSON{
			"productCode": productCode,
		},
	}
	if lang != "" {
		req.CustomData["lang"] = lang
	}
	if orderType == types.OrderTypeNew {
		req.Items = []types.OrderItemRequest{{
			ID:         uuid.NewString(),
			ExternalID: &productCode,
		}}
	}

	_, err := s.intTransport.Send("order/create", req, types.SendOptions{Timeout: 3 * time.Second})
	if err != nil {
		return errs.WrapWithFields(
			fmt.Errorf("send order/create failed %w", err),
			map[string]interface{}{
				"orderId":     req.ID,
				"phoneNum":    phoneNum,
				"productCode": productCode,
				"orderType":   orderType,
			},
		)
	}
	return nil
}

func (s *Viva) completeOrder(order Order) error {
	if order.Status == "error" {
		return nil
	}
	if len(order.Items) == 0 {
		return errs.WrapWithFields(
			fmt.Errorf("order items is empty"),
			map[string]interface{}{"orderId": order.ID},
		)
	}

	item := order.Items[0]
	for _, it := range order.Items {
		if it.Type == "activate" || it.Type == "reactivate" {
			item = it
			break
		}
	}
	if item.Type != "activate" && item.Type != "reactivate" {
		return nil
	}

	code := strings.TrimSpace(cast.ToString(item.Artifacts["ActivationCode"]))
	if code == "" {
		return errs.WrapWithFields(
			fmt.Errorf("activation code is empty"),
			map[string]interface{}{"orderId": order.ID, "itemId": item.ID},
		)
	}

	rawDownload, _ := item.Artifacts["download"].([]interface{})
	var downloadURL string
	for _, v := range rawDownload {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if downloadURL = strings.TrimSpace(cast.ToString(m["url"])); downloadURL != "" {
			break
		}
	}
	if downloadURL == "" {
		return errs.WrapWithFields(
			fmt.Errorf("download url not found"),
			map[string]interface{}{"orderId": order.ID, "itemId": item.ID},
		)
	}

	phone := strings.TrimSpace(cast.ToString(order.Fields["phone"]))
	if phone == "" {
		return errs.WrapWithFields(
			fmt.Errorf("order has no phone"),
			map[string]interface{}{"orderId": order.ID},
		)
	}

	productCode := productCodeFromOrder(order, item)
	product, err := s.catalog.GetProductByExternalId(productCode)
	if err != nil {
		return err
	}

	lang := strings.TrimSpace(cast.ToString(order.CustomData["lang"]))
	if lang == "" {
		lang = s.storedLang(phone)
	}
	if lang == "" {
		lang = "ru"
	}

	data := map[string]interface{}{
		"Phone":          phone,
		"ExternalID":     productCode,
		"ProductName":    cast.ToString(item.Product.Name),
		"Quantity":       cast.ToInt(item.Options["quantity"]),
		"ActivationCode": code,
		"DownloadURL":    downloadURL,
		"Link":           downloadURL,
		"Language":       lang,
	}

	if text := product.GetNotify("welcome_trial", data, lang); text != "" {
		if err := s.notify(phone, text); err != nil {
			return err
		}
	}

	text := product.GetNotify("license", data, lang)
	if text == "" {
		return errs.WrapWithFields(
			fmt.Errorf("license notification template is empty"),
			map[string]interface{}{
				"orderId":     order.ID,
				"productCode": productCode,
				"lang":        lang,
			},
		)
	}

	return s.notify(phone, text)
}

func productCodeFromOrder(order Order, item types.OrderItemResponse) string {
	if order.CustomData != nil {
		if code := strings.TrimSpace(cast.ToString(order.CustomData["productCode"])); code != "" {
			return code
		}
	}
	if len(item.Tariff.ExternalID) > 0 {
		if code := strings.TrimSpace(item.Tariff.ExternalID[0]); code != "" {
			return code
		}
	}
	return strings.TrimSpace(item.Tariff.VendorCode)
}
