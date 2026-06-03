package viva_api

import (
	"strings"
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func sampleCompletedOrder() types.OrderResponse {
	return types.OrderResponse{
		ID:     "5a10cc13-0983-5ffd-9b4d-8bef51df0dfb",
		Status: string(types.OrderStatusListCompleted),
		Fields: map[string]interface{}{"phone": "37477600552"},
		CustomData: map[string]interface{}{
			"productCode": "SAFEKID",
			"lang":        "ru",
		},
		Items: []types.OrderItemResponse{{
			Type: "activate",
			Artifacts: map[string]interface{}{
				"ActivationCode": "LICENSE-KEY-42",
				"download": []interface{}{
					map[string]interface{}{"url": "https://download.example/safekids"},
				},
			},
		}},
	}
}

func TestOrderCompleteHandler_Steps11To13_SendsWelcomeAndLicenseSMS(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
	}

	v.orderCompleteHandler(&orderCompleteCtx{order: sampleCompletedOrder()})

	if len(notifyTr.sendCalls) != 2 {
		t.Fatalf("expected 2 outbound SMS, got %d", len(notifyTr.sendCalls))
	}
	welcome := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	license := notifyTr.sendCalls[1].msg.(map[string]interface{})["text"].(string)
	if !strings.Contains(welcome, "Поздравляем! Услуга Kaspersky Safe Kids подключена") {
		t.Fatalf("welcome sms = %q", welcome)
	}
	if !strings.Contains(welcome, "https://download.example/safekids") {
		t.Fatalf("welcome sms missing link: %q", welcome)
	}
	if license != "Kaspersky Safe Kids: LICENSE-KEY-42" {
		t.Fatalf("license sms = %q", license)
	}
	to := notifyTr.sendCalls[0].msg.(map[string]interface{})["to"].(string)
	if to != "37477600552" {
		t.Fatalf("to = %q", to)
	}
}

func TestCompleteOrder_SkipsWhenNoActivateItem(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
	}
	order := sampleCompletedOrder()
	order.Items[0].Type = "renew"

	if err := v.completeOrder(order); err != nil {
		t.Fatal(err)
	}
	if len(notifyTr.sendCalls) != 0 {
		t.Fatalf("expected no SMS, got %d", len(notifyTr.sendCalls))
	}
}

type orderCompleteCtx struct {
	order types.OrderResponse
}

func (c *orderCompleteCtx) Data(dest interface{}) {
	*dest.(*types.OrderResponse) = c.order
}

func (c *orderCompleteCtx) Param(string) string          { return "" }
func (c *orderCompleteCtx) Headers(interface{})        {}
func (c *orderCompleteCtx) Response(interface{}) error { return nil }
func (c *orderCompleteCtx) Error(error) error          { return nil }
