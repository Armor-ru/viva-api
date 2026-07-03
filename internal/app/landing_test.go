package viva_api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
)

func TestLandingErrorPayloadWithResultCode(t *testing.T) {
	err := errs.WrapWithFields(
		errors.New("not enough funds"),
		map[string]interface{}{"resultCode": 14},
	)

	payload := landingErrorPayload(err)
	if payload.Error != "not enough funds" {
		t.Fatalf("Error = %q, want %q", payload.Error, "not enough funds")
	}
	if payload.ResultCode == nil || *payload.ResultCode != 14 {
		t.Fatalf("ResultCode = %v, want 14", payload.ResultCode)
	}
}

func TestLandingErrorPayloadWithoutResultCode(t *testing.T) {
	payload := landingErrorPayload(errors.New("viva unavailable"))
	if payload.Error != "viva unavailable" {
		t.Fatalf("Error = %q, want %q", payload.Error, "viva unavailable")
	}
	if payload.ResultCode != nil {
		t.Fatalf("ResultCode = %v, want nil", *payload.ResultCode)
	}
}

func TestLandingHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		resultCode int
		want       int
	}{
		{name: "not allowed product", resultCode: 1, want: http.StatusForbidden},
		{name: "product not found", resultCode: 2, want: http.StatusNotFound},
		{name: "already active", resultCode: 7, want: http.StatusConflict},
		{name: "invalid phone", resultCode: 12, want: http.StatusBadRequest},
		{name: "not enough funds", resultCode: 14, want: http.StatusUnprocessableEntity},
		{name: "no pending subscription", resultCode: 17, want: http.StatusConflict},
		{name: "not verified", resultCode: 19, want: http.StatusUnprocessableEntity},
		{name: "sms not sent", resultCode: 21, want: http.StatusBadGateway},
		{name: "too many requests", resultCode: 23, want: http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errs.WrapWithFields(
				errors.New("viva rejected request"),
				map[string]interface{}{"resultCode": tt.resultCode},
			)
			if got := landingHTTPStatus(http.StatusBadGateway, err); got != tt.want {
				t.Fatalf("landingHTTPStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLandingHTTPStatusWithoutResultCode(t *testing.T) {
	if got := landingHTTPStatus(http.StatusBadRequest, errors.New("invalid request")); got != http.StatusBadRequest {
		t.Fatalf("landingHTTPStatus() = %d, want %d", got, http.StatusBadRequest)
	}
}

func TestLandingHTTPStatusTimeout(t *testing.T) {
	err := fmt.Errorf("confirm subscription failed: %w", context.DeadlineExceeded)
	if got := landingHTTPStatus(http.StatusBadGateway, err); got != http.StatusGatewayTimeout {
		t.Fatalf("landingHTTPStatus() = %d, want %d", got, http.StatusGatewayTimeout)
	}
}
