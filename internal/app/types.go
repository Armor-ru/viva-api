package viva_api

import "time"

// SmppConfig используется в интеграционных тестах (прямая отправка SMPP).
type SmppAddressConfig struct {
	SourceAddr    string `yaml:"sourceAddr"`
	SourceAddrTON uint8  `yaml:"sourceAddrTON"`
	SourceAddrNPI uint8  `yaml:"sourceAddrNPI"`
	DestAddrTON   uint8  `yaml:"destAddrTON"`
	DestAddrNPI   uint8  `yaml:"destAddrNPI"`
}

type SmppConfig struct {
	Endpoint []string          `yaml:"endpoint"`
	Auth     struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
	Address SmppAddressConfig `yaml:"address"`
}

type AppConfig struct {
	CatalogDir          string `yaml:"catalogDir"`
	DefaultLanguage     string `yaml:"defaultLanguage"`
	LangPreferenceTTL   string `yaml:"langPreferenceTTL"`
	LandingConfirmURL   string `yaml:"landingConfirmURL"`
}

func (a AppConfig) ResolvedDefaultLanguage() string {
	if a.DefaultLanguage != "" {
		return a.DefaultLanguage
	}
	return "ru"
}

func (a AppConfig) ResolvedLangPreferenceTTL() time.Duration {
	if a.LangPreferenceTTL == "" {
		return defaultLangPreferenceTTL
	}
	d, err := time.ParseDuration(a.LangPreferenceTTL)
	if err != nil || d <= 0 {
		return defaultLangPreferenceTTL
	}
	return d
}

type ExtReq struct {
	PhoneNum    string `json:"phoneNum" validate:"required"`
	ProductCode string `json:"productCode" validate:"required"`
}
