package viva_api

import (
	"encoding/json"
	"testing"
)

func TestInboundSMSDTO_UnmarshalNumericDestinationAddr(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"sourceAddr":"37477600552","destinationAddr":1020,"shortMessage":"1"}`)
	var sms inboundSMSDTO
	if err := json.Unmarshal(raw, &sms); err != nil {
		t.Fatal(err)
	}
	mo := parseInboundMO(sms)
	if !mo.isActivationOne() {
		t.Fatalf("mo = %+v", mo)
	}
}
