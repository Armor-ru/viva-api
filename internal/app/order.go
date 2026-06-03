package viva_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/tplext"
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
	data, err := json.Marshal(raw)
	if err != nil {
		return Order{}, err
	}
	if err := json.Unmarshal(data, &order); err != nil {
		return Order{}, fmt.Errorf("decode order/get response: %w", err)
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
		return fmt.Errorf("intTransport is not configured")
	}

	phoneNum = strings.TrimSpace(phoneNum)
	productCode = strings.TrimSpace(productCode)
	if phoneNum == "" || productCode == "" {
		return fmt.Errorf("phoneNum and productCode are required")
	}

	lang = strings.TrimSpace(lang)
	if lang == "" && s.langStore != nil {
		lang = strings.TrimSpace(s.langStore[phoneNum])
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
		return fmt.Errorf("send order/create: %w", err)
	}
	return nil
}

func (s *Viva) completeOrder(order Order) error {
	if order.Status == "error" {
		return nil
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order items is empty")
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
		return fmt.Errorf("ActivationCode is empty")
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
		return fmt.Errorf("download url not found")
	}

	phone := strings.TrimSpace(cast.ToString(order.Fields["phone"]))
	if phone == "" {
		return fmt.Errorf("order has no phone")
	}

	smsData := SmsData{
		ProductName:    cast.ToString(item.Product.Name),
		Quantity:       cast.ToInt(item.Options["quantity"]),
		ActivationCode: code,
		DownloadURL:    downloadURL,
	}

	var buf bytes.Buffer
	tpl, _ := template.New("sms").Funcs(tplext.Funcs).Parse(
		"{{.ProductName}}\nActivation code: {{.ActivationCode}}\nDownload link: {{.DownloadURL}}",
	)
	if err := tpl.Execute(&buf, smsData); err != nil {
		return fmt.Errorf("render sms template: %w", err)
	}
	return s.notify(phone, buf.String())
}
