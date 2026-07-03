package viva_api

type UssdRequest struct {
	Phone       string `json:"sourceAddr" validate:"required"`
	ShortNumber string `json:"destinationAddr" validate:"required"`
	Text        string `json:"shortMessage" validate:"required"`
}

type ExtReq struct {
	PhoneNum    string `json:"phoneNum" validate:"required"`
	ProductCode string `json:"productCode" validate:"required"`
}

type LandingErrorResponse struct {
	Error      string `json:"error"`
	ResultCode *int   `json:"resultCode,omitempty"`
}
