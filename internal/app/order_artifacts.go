package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"github.com/spf13/cast"
)

// activationArtifacts извлекается из artifacts позиции order/completed (SDS / provider-kss).
type activationArtifacts struct {
	ActivationCode string
	DownloadURL    string
}

func firstActivateOrderItem(items []types.OrderItemResponse) (types.OrderItemResponse, bool) {
	for _, it := range items {
		switch it.Type {
		case "activate", "reactivate":
			return it, true
		}
	}
	return types.OrderItemResponse{}, false
}

func activationArtifactsFromItem(item types.OrderItemResponse) (activationArtifacts, error) {
	if item.Artifacts == nil {
		return activationArtifacts{}, fmt.Errorf("order item artifacts are empty")
	}

	code := strings.TrimSpace(cast.ToString(item.Artifacts["ActivationCode"]))
	if code == "" {
		return activationArtifacts{}, fmt.Errorf("ActivationCode is empty in order/completed artifacts")
	}

	downloadURL, err := firstDownloadURL(item.Artifacts["download"])
	if err != nil {
		return activationArtifacts{}, err
	}

	return activationArtifacts{
		ActivationCode: code,
		DownloadURL:    downloadURL,
	}, nil
}
