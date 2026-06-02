package viva_api

type ChannelMap struct {
	Id    string `yaml:"id"`
	MapId string `yaml:"mapId"`
}

type Channels struct {
	DefaultMapId string       `yaml:"defaultMapId"`
	Map          []ChannelMap `yaml:"map"`
}

type SmppAddressConfig struct {
	SourceAddr    string `yaml:"sourceAddr" json:"sourceAddr"`
	SourceAddrTON uint8  `yaml:"sourceAddrTON" json:"sourceAddrTON"`
	SourceAddrNPI uint8  `yaml:"sourceAddrNPI" json:"sourceAddrNPI"`
	DestAddrTON   uint8  `yaml:"destAddrTON" json:"destAddrTON"`
	DestAddrNPI   uint8  `yaml:"destAddrNPI" json:"destAddrNPI"`
}

type SmppConfig struct {
	Endpoint []string `yaml:"endpoint" json:"endpoint"`
	Auth     struct {
		User     string `yaml:"user" json:"user"`
		Password string `yaml:"password" json:"password"`
	} `yaml:"auth" json:"auth"`
	Address SmppAddressConfig `yaml:"address" json:"address"`
}

type SmsConfig struct {
	MenuDir         string `yaml:"menuDir"`
	DefaultLanguage string `yaml:"defaultLanguage"`
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
