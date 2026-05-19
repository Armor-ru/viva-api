package vivaclient

type ResponseModel struct {
	ResultCode int         `json:"resultCode"`
	Message    interface{} `json:"message"`
	Result     bool        `json:"result"`
}
