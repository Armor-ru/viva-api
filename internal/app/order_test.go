package viva_api

import (
	"testing"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestGetOrderId_Deterministic(t *testing.T) {
	t.Parallel()

	v := &viva{accountId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}
	a := v.getOrderId("37477600552", "SAFEKID")
	b := v.getOrderId("37477600552", "SAFEKID")
	if a != b || a == "" {
		t.Fatalf("getOrderId() = %q", a)
	}
}

func TestIsActiveOrder(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)
	v := &viva{}

	if !v.isActiveOrder(Order{Status: string(types.OrderStatusListCompleted), EndTime: &future}) {
		t.Fatal("expected active")
	}
	if v.isActiveOrder(Order{Status: string(types.OrderStatusListCompleted), EndTime: &past}) {
		t.Fatal("expected expired")
	}
}
