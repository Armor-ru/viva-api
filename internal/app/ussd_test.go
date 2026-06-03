package viva_api

import (
	"os"
	"path/filepath"
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

func writeTestCatalog(t *testing.T, dir string) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
}

func loadTestCatalog(t *testing.T) catalog.Catalog {
	t.Helper()
	dir := t.TempDir()
	writeTestCatalog(t, dir)
	store := catalog.NewCatalog()
	if err := store.Load(dir); err != nil {
		t.Fatal(err)
	}
	_ = store.SetDefaultLang("ru")
	return store
}

func TestHandleActivationOne_Step2_FindsProduct(t *testing.T) {
	t.Parallel()

	v := &viva{
		catalog:   loadTestCatalog(t),
		accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}
	v.handleActivationOne(inboundMO{
		Phone: "37477600552", ShortNumber: "1020", Text: "1",
	})
}

func TestHandleActivationOne_Step2_UnknownShortNumber(t *testing.T) {
	t.Parallel()

	v := &viva{
		catalog: loadTestCatalog(t),
	}
	v.handleActivationOne(inboundMO{
		Phone: "37477600552", ShortNumber: "9999", Text: "1",
	})
}

func TestHandleActivationOne_Step3_OrderId(t *testing.T) {
	t.Parallel()

	const accountID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	v := &viva{
		catalog:   loadTestCatalog(t),
		accountId: accountID,
	}
	phone := "37477600552"
	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})

	want := v.getOrderId(phone, "SAFEKID")
	got := v.getOrderId(phone, "SAFEKID")
	if want != got || want == "" {
		t.Fatalf("orderId = %q", got)
	}
}

type orderGetTransport struct {
	fakeTransport
	getErr error
	order  Order
}

func (f *orderGetTransport) Send(topic string, msg types.Message, opt types.SendOptions) (types.Message, error) {
	f.sendCalls = append(f.sendCalls, sendCall{topic: topic, msg: msg})
	if topic == "order/get" {
		if f.getErr != nil {
			return nil, f.getErr
		}
		return f.order, nil
	}
	return map[string]interface{}{}, nil
}

func TestHandleActivationOne_Step4_PublishesOrderGet(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	v := &viva{accountId: accountID}
	orderID := v.getOrderId(phone, "SAFEKID")

	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	v = &viva{
		intTransport: intTr,
		catalog:      loadTestCatalog(t),
		accountId:    accountID,
	}

	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})

	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("expected order/get, got %+v", intTr.sendCalls)
	}
	req, ok := intTr.sendCalls[0].msg.(types.GetOrderRequest)
	if !ok || req.ID != orderID {
		t.Fatalf("request = %+v, want id %q", intTr.sendCalls[0].msg, orderID)
	}
}

func TestUssdHandler_ActivationOne_OrderGetVivaInit_LandingSMS(t *testing.T) {
	t.Parallel()

	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	vivaAPI := newVivaAPITestDouble(t)
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:      intTr,
		vivaClient:        vivaAPI.Client(),
		ussdTransport:   notifyTr,
		catalog:           loadTestCatalog(t),
		accountId:         "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		langStore:         NewLangStore("ru"),
		landingConfirmURL: "https://landing.test/confirm",
	}

	v.ussdHandler(&handlerCtx{payload: inboundSMSDTO{
		Phone: "37477600552", ShortNumber: "1020", Text: "1",
	}})

	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("expected order/get only, got %+v", intTr.sendCalls)
	}
	if vivaAPI.InitCalls != 1 || vivaAPI.ConfirmCalls != 0 {
		t.Fatalf("expected Viva init only, got init=%d confirm=%d", vivaAPI.InitCalls, vivaAPI.ConfirmCalls)
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected landing link SMS, got %d", len(notifyTr.sendCalls))
	}
}

type errOrderNotFoundSentinel struct{}

func (errOrderNotFoundSentinel) Error() string { return "order not found" }

func errOrderNotFound() error { return errOrderNotFoundSentinel{} }

func TestSmppInbound_Stop_UsesStopHandlerNotLegacyRules(t *testing.T) {
	t.Parallel()

	vivaAPI := newVivaAPITestDouble(t)
	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	v := &viva{
		intTransport: intTr,
		vivaClient:   vivaAPI.Client(),
		catalog:      loadTestCatalog(t),
		accountId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	v.ussdHandler(&handlerCtx{payload: inboundSMSDTO{
		Phone: "37477600552", ShortNumber: "1020", Text: "STOP",
	}})

	if vivaAPI.RemoveCalls != 1 {
		t.Fatalf("expected Viva RemoveSubscription, got %d", vivaAPI.RemoveCalls)
	}
}

type handlerCtx struct {
	payload interface{}
}

func (c *handlerCtx) Data(dest interface{}) {
	*dest.(*inboundSMSDTO) = c.payload.(inboundSMSDTO)
}

func (c *handlerCtx) Param(string) string              { return "" }
func (c *handlerCtx) Headers(interface{})            {}
func (c *handlerCtx) Response(interface{}) error     { return nil }
func (c *handlerCtx) Error(error) error              { return nil }
