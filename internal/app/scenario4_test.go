package viva_api

import (
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestScenario4_Step1_AcceptsValidWebhook(t *testing.T) {
	t.Parallel()

	err := scenario4Step1(ExtReq{PhoneNum: "37477600552", ProductCode: "SAFEKID"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestScenario4_Step1_RejectsEmptyPhone(t *testing.T) {
	t.Parallel()

	if err := scenario4Step1(ExtReq{ProductCode: "SAFEKID"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario4_Step1_RejectsEmptyProductCode(t *testing.T) {
	t.Parallel()

	if err := scenario4Step1(ExtReq{PhoneNum: "37477600552"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario4_Step3_ProductByExternalId(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.productForExternalID(FlowPaid, "3", "SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" || product.GetShortNumber() != "1020" {
		t.Fatalf("product = %+v", product)
	}
}

func TestScenario4_Step3_UnknownExternalId(t *testing.T) {
	t.Parallel()

	intTr := &fakeTransport{}
	v := &viva{
		intTransport: intTr,
		catalog:      loadTestCatalog(t),
		accountId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	v.handleExtAppPartnerProductActivation(&webhookCtx{
		body: ExtReq{PhoneNum: "37477600552", ProductCode: "UNKNOWN"},
	})

	if len(intTr.sendCalls) != 0 {
		t.Fatalf("expected no order/create, got %d", len(intTr.sendCalls))
	}
}

func TestScenario4_Step4_OrderIdFromProductAndPhone(t *testing.T) {
	t.Parallel()

	v := &viva{
		accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		catalog:   loadTestCatalog(t),
	}
	product, err := v.productForExternalID(FlowPaid, "3", "SAFEKID")
	if err != nil {
		t.Fatal(err)
	}

	orderID, err := v.scenario4Step4FormOrderID("37477600552", product)
	if err != nil {
		t.Fatal(err)
	}
	want := v.getOrderId("37477600552", "SAFEKID")
	if orderID != want || orderID == "" {
		t.Fatalf("orderId = %q want %q", orderID, want)
	}
}

func TestScenario4_Step5_BuildsRenewOrderCreate(t *testing.T) {
	t.Parallel()

	v := &viva{
		accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		langStore: NewLangStore("ru"),
	}
	orderID := v.getOrderId("37477600552", "SAFEKID")
	req := v.scenario4Step5ConvertToRenewOrderCreate(orderID, "37477600552", "SAFEKID")

	if req.Type != types.OrderTypeRenew {
		t.Fatalf("type = %q", req.Type)
	}
	if req.ID != orderID {
		t.Fatalf("id = %q", req.ID)
	}
	if req.Fields["phone"] != "37477600552" {
		t.Fatalf("phone field = %v", req.Fields["phone"])
	}
	if req.CustomData["productCode"] != "SAFEKID" {
		t.Fatalf("productCode = %v", req.CustomData["productCode"])
	}
	if len(req.Items) != 0 {
		t.Fatalf("renew should have no items, got %d", len(req.Items))
	}
}

func sampleRenewCompletedOrder() types.OrderResponse {
	return types.OrderResponse{
		ID:     "5a10cc13-0983-5ffd-9b4d-8bef51df0dfb",
		Status: string(types.OrderStatusListCompleted),
		Fields: map[string]interface{}{"phone": "37477600552"},
		CustomData: map[string]interface{}{
			"productCode": "SAFEKID",
			"lang":        "ru",
		},
		Items: []types.OrderItemResponse{{Type: "renew"}},
	}
}

func TestScenario4_Step9_ReceivesOrderCompleted(t *testing.T) {
	t.Parallel()

	err := scenario4Step9ReceiveOrderCompleted(sampleRenewCompletedOrder())
	if err != nil {
		t.Fatal(err)
	}
}

func TestScenario4_Step10_ProductByExternalId(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	product, err := v.productForExternalID(FlowPaid, "10", "SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" || product.GetShortNumber() != "1020" {
		t.Fatalf("product = %+v", product)
	}
}

func TestScenario4_Step10_FromOrderCompleted(t *testing.T) {
	t.Parallel()

	v := &viva{catalog: loadTestCatalog(t)}
	externalID := productCodeFromOrderResponse(sampleRenewCompletedOrder())
	product, err := v.productForExternalID(FlowPaid, "10", externalID)
	if err != nil {
		t.Fatal(err)
	}
	if product.GetExternalID() != "SAFEKID" {
		t.Fatalf("externalId = %q", product.GetExternalID())
	}
}

func TestScenario4_Step10_UnknownProductOnOrderCompleted(t *testing.T) {
	t.Parallel()

	order := sampleRenewCompletedOrder()
	order.CustomData["productCode"] = "UNKNOWN"
	v := &viva{catalog: loadTestCatalog(t)}
	if err := v.handleScenario4OrderCompleted(order); err == nil {
		t.Fatal("expected error")
	}
}

func TestScenario4_Step11_SendsSMS4_Russian(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
		paidWelcomeSent: NewPaidWelcomeStore(),
	}

	if err := v.handleScenario4OrderCompleted(sampleRenewCompletedOrder()); err != nil {
		t.Fatal(err)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("sms count = %d", len(notifyTr.sendCalls))
	}
	text := notifyTr.sendCalls[0].msg.(map[string]interface{})["text"].(string)
	if notifyTr.sendCalls[0].msg.(map[string]interface{})["to"].(string) != "37477600552" {
		t.Fatal("wrong recipient")
	}
	wantSub := "Поздравляем! Платная версия услуги Kaspersky Safe Kids подключена"
	if len(text) < len(wantSub) || text[:len(wantSub)] != wantSub {
		t.Fatalf("sms #4 = %q", text)
	}
}

func TestScenario4_Step12_SkipsSMS4OnSecondRenew(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		langStore:       NewLangStore("ru"),
		paidWelcomeSent: NewPaidWelcomeStore(),
	}
	order := sampleRenewCompletedOrder()

	if err := v.handleScenario4OrderCompleted(order); err != nil {
		t.Fatal(err)
	}
	if err := v.handleScenario4OrderCompleted(order); err != nil {
		t.Fatal(err)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected 1 SMS (trial→paid once), got %d", len(notifyTr.sendCalls))
	}
}

func TestScenario4_Step9_HandlerRoutesRenewOrder(t *testing.T) {
	t.Parallel()

	notifyTr := &fakeTransport{}
	v := &viva{
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		paidWelcomeSent: NewPaidWelcomeStore(),
	}
	v.orderCompleteHandler(&orderCompleteCtx{order: sampleRenewCompletedOrder()})

	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected 1 SMS #4, got %d", len(notifyTr.sendCalls))
	}
}

func TestScenario4_Step6_PublishesOrderCreateToNATS(t *testing.T) {
	t.Parallel()

	intTr := &fakeTransport{}
	v := &viva{
		intTransport: intTr,
		accountId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}
	orderID := v.getOrderId("37477600552", "SAFEKID")
	req := v.scenario4Step5ConvertToRenewOrderCreate(orderID, "37477600552", "SAFEKID")

	if err := v.scenario4Step6PublishOrderCreate(req); err != nil {
		t.Fatal(err)
	}
	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/create" {
		t.Fatalf("calls = %+v", intTr.sendCalls)
	}
}

func TestScenario4_Step2_PublishesOrderCreateRenew(t *testing.T) {
	t.Parallel()

	intTr := &fakeTransport{}
	v := &viva{
		intTransport: intTr,
		catalog:      loadTestCatalog(t),
		accountId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	v.handleExtAppPartnerProductActivation(&webhookCtx{
		body: ExtReq{PhoneNum: "37477600552", ProductCode: "SAFEKID"},
	})

	if len(intTr.sendCalls) != 1 {
		t.Fatalf("expected 1 NATS send, got %d", len(intTr.sendCalls))
	}
	if intTr.sendCalls[0].topic != "order/create" {
		t.Fatalf("topic = %q", intTr.sendCalls[0].topic)
	}
	req, ok := intTr.sendCalls[0].msg.(types.OrderCreateRequest)
	if !ok {
		t.Fatalf("msg type %T", intTr.sendCalls[0].msg)
	}
	if req.Type != types.OrderTypeRenew {
		t.Fatalf("order type = %q", req.Type)
	}
	wantID := v.getOrderId("37477600552", "SAFEKID")
	if req.ID != wantID {
		t.Fatalf("order id = %q want %q", req.ID, wantID)
	}
}

func TestScenario4_Step2_SkipsOrderCreateOnInvalidPayload(t *testing.T) {
	t.Parallel()

	intTr := &fakeTransport{}
	v := &viva{intTransport: intTr, accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}

	v.handleExtAppPartnerProductActivation(&webhookCtx{body: ExtReq{ProductCode: "SAFEKID"}})

	if len(intTr.sendCalls) != 0 {
		t.Fatalf("expected no NATS send, got %d", len(intTr.sendCalls))
	}
}

type webhookCtx struct {
	body     ExtReq
	response interface{}
}

func (c *webhookCtx) Data(dest interface{}) {
	switch d := dest.(type) {
	case *ExtReq:
		*d = c.body
	}
}

func (c *webhookCtx) Param(string) string { return "" }
func (c *webhookCtx) Headers(interface{})  {}
func (c *webhookCtx) Response(v interface{}) error {
	c.response = v
	return nil
}
func (c *webhookCtx) Error(error) error { return nil }
