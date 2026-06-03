package viva_api

import (
	"fmt"
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

func TestActivationArtifactsFromOrder_Step11(t *testing.T) {
	t.Parallel()

	got, err := artifactsFromCompletedOrder(sampleCompletedOrder())
	if err != nil {
		t.Fatal(err)
	}
	if got.ActivationCode != "LICENSE-KEY-42" {
		t.Fatalf("ActivationCode = %q", got.ActivationCode)
	}
	if got.DownloadURL != "https://download.example/safekids" {
		t.Fatalf("DownloadURL = %q", got.DownloadURL)
	}
}

func TestActivationArtifactsFromOrder_PicksActivateItemNotFirst(t *testing.T) {
	t.Parallel()

	order := sampleCompletedOrder()
	order.Items = []types.OrderItemResponse{
		{Type: "renew", Artifacts: map[string]interface{}{"ActivationCode": "WRONG"}},
		{
			Type: "activate",
			Artifacts: map[string]interface{}{
				"ActivationCode": "CORRECT",
				"download": []interface{}{
					map[string]interface{}{"url": "https://example.com/app"},
				},
			},
		},
	}

	got, err := artifactsFromCompletedOrder(order)
	if err != nil {
		t.Fatal(err)
	}
	if got.ActivationCode != "CORRECT" || got.DownloadURL != "https://example.com/app" {
		t.Fatalf("got = %+v", got)
	}
}

func TestActivationArtifactsFromOrder_MissingCode(t *testing.T) {
	t.Parallel()

	order := sampleCompletedOrder()
	order.Items[0].Artifacts = map[string]interface{}{
		"download": []interface{}{
			map[string]interface{}{"url": "https://example.com"},
		},
	}
	if _, err := artifactsFromCompletedOrder(order); err == nil {
		t.Fatal("expected error for missing ActivationCode")
	}
}

func artifactsFromCompletedOrder(order types.OrderResponse) (activationArtifacts, error) {
	item, ok := firstActivateOrderItem(order.Items)
	if !ok {
		return activationArtifacts{}, fmt.Errorf("order has no activate or reactivate item")
	}
	return activationArtifactsFromItem(item)
}

func TestActivationArtifactsFromOrder_MissingDownload(t *testing.T) {
	t.Parallel()

	order := sampleCompletedOrder()
	order.Items[0].Artifacts = map[string]interface{}{
		"ActivationCode": "KEY",
	}
	if _, err := artifactsFromCompletedOrder(order); err == nil {
		t.Fatal("expected error for missing download url")
	}
}
