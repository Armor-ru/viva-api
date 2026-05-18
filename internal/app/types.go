package viva_api

type ChannelMap struct {
	Id    string `yaml:"id"`
	MapId string `yaml:"mapId"`
}

type Channels struct {
	DefaultMapId string       `yaml:"defaultMapId"`
	Map          []ChannelMap `yaml:"map"`
}

type SmppConfig struct {
	Endpoint []string `yaml:"endpoint" json:"endpoint"`
	Auth     struct {
		User     string `yaml:"user" json:"user"`
		Password string `yaml:"password" json:"password"`
	} `yaml:"auth" json:"auth"`
	Template string `yaml:"template" json:"template"`
}

type ExtReq struct {
	PhoneNum    string `json:"phoneNum" validate:"required"`
	ProductCode string `json:"productCode" validate:"required"`
	SmsScenario string `json:"smsScenario,omitempty"`
	Locale      string `json:"locale,omitempty"`
}

type SmsData struct {
	ProductName    string
	ProductLabel   string
	Quantity       int
	ActivationCode string
	LicenseKey     string
	DownloadURL    string
	TrialEndDate   string
}

const (
	cdVivaWebhook = "vivaWebhook"
	cdSmsScenario = "smsScenario"
	cdSmsLocale   = "smsLocale"

	whActivationReq = "activationRequest"
	whActivation    = "activation"
	whRemove        = "remove"
)
