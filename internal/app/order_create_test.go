package viva_api

import (
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestCreateOrder_CustomDataIncludesLang(t *testing.T) {
	t.Parallel()

	intTr := &fakeTransport{}
	v := &viva{
		intTransport: intTr,
		accountId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		langStore:    NewLangStore("ru"),
	}

	if err := v.createOrder(types.OrderTypeNew, "37477600552", "SAFEKID", "ru"); err != nil {
		t.Fatal(err)
	}
	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/create" {
		t.Fatalf("calls = %+v", intTr.sendCalls)
	}

	req, ok := intTr.sendCalls[0].msg.(types.OrderCreateRequest)
	if !ok {
		t.Fatalf("payload type %T", intTr.sendCalls[0].msg)
	}
	if req.CustomData["productCode"] != "SAFEKID" {
		t.Fatalf("productCode = %v", req.CustomData["productCode"])
	}
	if req.CustomData["lang"] != "ru" {
		t.Fatalf("lang = %v", req.CustomData["lang"])
	}
	if req.Fields["phone"] != "37477600552" {
		t.Fatalf("phone = %v", req.Fields["phone"])
	}
	if req.Type != types.OrderTypeNew {
		t.Fatalf("type = %q, want %q", req.Type, types.OrderTypeNew)
	}
	if len(req.Items) != 1 || req.Items[0].ExternalID == nil || *req.Items[0].ExternalID != "SAFEKID" {
		t.Fatalf("items = %+v", req.Items)
	}
}

func TestHandleActivationOne_NoOrderCreateOnMO(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	vivaAPI := newVivaAPITestDouble(t)
	notifyTr := &fakeTransport{}
	v := &viva{
		intTransport:      intTr,
		ussdTransport:   notifyTr,
		vivaClient:        vivaAPI.Client(),
		catalog:           loadTestCatalog(t),
		accountId:         accountID,
		langStore:         NewLangStore("ru"),
		landingConfirmURL: "https://landing.test/confirm",
	}

	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})

	if len(intTr.sendCalls) != 1 || intTr.sendCalls[0].topic != "order/get" {
		t.Fatalf("expected order/get only, got %+v", intTr.sendCalls)
	}
}
