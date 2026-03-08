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
}

type SmsData struct {
	ProductName    string
	Quantity       int
	ActivationCode string
	DownloadURL    string
}
