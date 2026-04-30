package vivaclient

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetSubInfoResponse — GET /api/Subscriber/{msisdn}
type GetSubInfoResponse struct {
	Msisdn            string  `json:"msisdn"`
	SubNo             int32   `json:"subNo"`
	AvailableProducts []int32 `json:"availableProducts"`
}

// ResponseModel — Init / Confirm / Remove subscription
type ResponseModel struct {
	ResultCode int         `json:"resultCode"`
	Message    interface{} `json:"message"`
	Result     bool        `json:"result"`
}

type ExtAppProductDTO struct {
	PhoneNumber string `json:"phoneNumber"`
	ProductName string `json:"productName"`
	ValidDays   int    `json:"validDays"`
	ExpDate     string `json:"expDate"`
}

type ExtAppProductsRespDTO struct {
	PhoneNumber              string             `json:"phoneNumber"`
	ActiveSubscribedProducts []ExtAppProductDTO `json:"activeSubscribedProducts"`
	AvailableProducts        []ExtAppProductDTO `json:"availableProducts"`
}

type extAppProductsEnvelope struct {
	ResultCode int                   `json:"resultCode"`
	Message    string                `json:"message"`
	Result     *ExtAppProductsRespDTO `json:"result"`
}
