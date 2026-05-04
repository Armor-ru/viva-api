package vivaclient

type GetSubInfoResponse struct {
	Msisdn            string  `json:"msisdn"`
	SubNo             int32   `json:"subNo"`
	AvailableProducts []int32 `json:"availableProducts"`
}

type ResponseModel struct {
	ResultCode int         `json:"resultCode"`
	Message    interface{} `json:"message"`
	Result     bool        `json:"result"`
}
