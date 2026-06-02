package viva_api

import (
	"errors"
	"testing"
)

func TestIsRetryableNotifyErr(t *testing.T) {
	if !isRetryableNotifyErr(errors.New("SMPP transport is not connected")) {
		t.Fatal("expected retryable")
	}
	if isRetryableNotifyErr(errors.New("SMPP text is empty")) {
		t.Fatal("expected non-retryable")
	}
}
