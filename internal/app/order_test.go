package viva_api

import (
	"testing"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestGetOrderId(t *testing.T) {
	t.Parallel()

	v := &Viva{accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}
	id1 := v.getOrderId("37477600552", "SAFEKID")
	id2 := v.getOrderId("37477600552", "SAFEKID")
	if id1 == "" || id1 != id2 {
		t.Fatalf("getOrderId() = %q", id1)
	}
}

func TestCreateOrder_SendsRequest(t *testing.T) {
	t.Parallel()

	tr := &fakeTransport{}
	v := &Viva{
		intTransport: tr,
		accountId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	if err := v.createOrder(types.OrderTypeNew, "37477600552", "SAFEKID", "en"); err != nil {
		t.Fatalf("createOrder() error = %v", err)
	}
	if len(tr.sendCalls) != 1 || tr.sendCalls[0].topic != "order/create" {
		t.Fatalf("unexpected send calls: %+v", tr.sendCalls)
	}
	req, ok := tr.sendCalls[0].msg.(types.OrderCreateRequest)
	if !ok {
		t.Fatalf("expected OrderCreateRequest, got %T", tr.sendCalls[0].msg)
	}
	if req.CustomData["lang"] != "en" {
		t.Fatalf("expected lang=en, got %v", req.CustomData["lang"])
	}
}

func TestIsActiveOrder(t *testing.T) {
	t.Parallel()

	v := &Viva{}
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	if v.isActiveOrder(Order{Status: string(types.OrderStatusListCompleted), EndTime: &future}) {
		// ok
	} else {
		t.Fatal("expected active order")
	}
	if v.isActiveOrder(Order{Status: string(types.OrderStatusListCompleted), EndTime: &past}) {
		t.Fatal("expected inactive order")
	}
	if v.isActiveOrder(Order{Status: "pending", EndTime: &future}) {
		t.Fatal("expected inactive order for non-completed status")
	}
}
