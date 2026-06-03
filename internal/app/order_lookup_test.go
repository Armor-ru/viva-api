package viva_api

import (
	"testing"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestLookupOrder_NotExists_OnError(t *testing.T) {
	t.Parallel()

	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	v := &viva{
		intTransport: intTr,
		accountId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}
	lookup := v.lookupOrder("any-id")
	if lookup.Exists {
		t.Fatal("expected not exists")
	}
}

func TestLookupOrder_Exists_Active(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(time.Hour)
	intTr := &orderGetTransport{
		order: Order{
			ID:      "oid-1",
			Status:  string(types.OrderStatusListCompleted),
			EndTime: &future,
		},
	}
	v := &viva{intTransport: intTr}
	lookup := v.lookupOrder("oid-1")
	if !lookup.Exists || !lookup.Active {
		t.Fatalf("lookup = %+v", lookup)
	}
}

func TestHandleActivationOne_Step5_OrderNotExists(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	intTr := &orderGetTransport{getErr: errOrderNotFound()}
	v := &viva{
		intTransport: intTr,
		catalog:      loadTestCatalog(t),
		accountId:    accountID,
	}
	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})
	if len(intTr.sendCalls) != 1 {
		t.Fatalf("sends = %d", len(intTr.sendCalls))
	}
}

func TestHandleActivationOne_Step5_OrderActive_RunsScenario2(t *testing.T) {
	t.Parallel()

	phone := "37477600552"
	accountID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	v := &viva{accountId: accountID}
	orderID := v.getOrderId(phone, "SAFEKID")
	future := time.Now().Add(time.Hour)
	intTr := &orderGetTransport{
		order: Order{ID: orderID, Status: string(types.OrderStatusListCompleted), EndTime: &future},
	}
	notifyTr := &fakeTransport{}
	v = &viva{
		intTransport:    intTr,
		ussdTransport: notifyTr,
		catalog:         loadTestCatalog(t),
		accountId:       accountID,
		langStore:       NewLangStore("ru"),
	}
	v.handleActivationOne(inboundMO{Phone: phone, ShortNumber: "1020", Text: "1"})
	if len(intTr.sendCalls) != 1 {
		t.Fatalf("expected order/get only, got %d", len(intTr.sendCalls))
	}
	if len(notifyTr.sendCalls) != 1 {
		t.Fatalf("expected scenario 2 SMS #8, got %d", len(notifyTr.sendCalls))
	}
}
