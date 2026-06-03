package viva_api

import (
	"encoding/json"
	"fmt"
)

// smppField принимает в JSON строку или число (SMPPSim может слать destinationAddr как 1020).
type smppField string

func (f *smppField) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = smppField(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("smpp field: expected string or number, got %s", string(data))
	}
	*f = smppField(n.String())
	return nil
}
