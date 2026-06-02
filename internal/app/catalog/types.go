package catalog

type Step struct {
	Command string                 `json:"command" yaml:"command"`
	Params  map[string]interface{} `json:"params" yaml:"params"`
}

type Product struct {
	ShortNumber     string              `json:"shortNumber" yaml:"shortNumber"`
	ExternalID      string              `json:"externalId" yaml:"externalId"`
	DefaultLanguage string              `json:"defaultLanguage" yaml:"defaultLanguage"`
	Scenarios       map[string][]Step   `json:"scenarios" yaml:"scenarios"`
	Notifications   map[string]map[string]string `json:"notifications" yaml:"notifications"`
}
