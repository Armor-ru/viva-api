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
	Endpoint     []string                       `yaml:"endpoint"`
	Auth         struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
	Template     string                         `yaml:"template"`
	SmsTemplates map[string]map[string]string   `yaml:"smsTemplates"`
}

type ExtReq struct {
	PhoneNum    string `json:"phoneNum" validate:"required"`
	ProductCode string `json:"productCode" validate:"required"`
	SmsScenario string `json:"smsScenario,omitempty"`
	Locale      string `json:"locale,omitempty"`
}

type SmsTplData struct {
	ProductName    string
	ProductLabel   string
	Quantity       int
	ActivationCode string
	LicenseKey     string
	DownloadURL    string
	TrialEndDate   string
}

const (
	CDVivaWebhook   = "vivaWebhook"
	CDSmsScenario   = "smsScenario"
	CDSmsLocale     = "smsLocale"
	WHActivationReq = "activationRequest"
	WHActivation    = "activation"
	WHRemove        = "remove"
)

const (
	scSms2  = "sms2"
	scSms3  = "sms3"
	scSms4  = "sms4"
	scSms5  = "sms5"
	scSms14 = "sms14"
	scSms15 = "sms15"
	scOff   = "sms_deactivated"
)
