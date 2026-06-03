package catalog

// Product — продукт каталога по SPEC
type Product interface {
	GetNotify(key string, data map[string]interface{}, lang ...string) string
	GetExternalID() string
	GetShortNumber() string
	GetDefaultLanguage() string
}

// product — запись продукта в каталоге (SPEC: ShortNumber, ExternalId, notifications).
type product struct {
	ShortNumber     string
	ExternalId      string
	DefaultLanguage string
	Notifications   map[string]map[string]string
	catalogLang     string
}

// Catalog — каталог продуктов по SPEC тимлида.
type Catalog interface {
	Load(dir string) error
	GetProductByShortNumber(id string) (Product, error)
	GetProductByExternalId(id string) (Product, error)
	SetDefaultLang(lang string) error
}

// catalog — файловый каталог (SPEC: products []Product).
type catalog struct {
	Products    []product
	byShort     map[string]*product
	byExternal  map[string]*product
	defaultLang string
}

var _ Catalog = (*catalog)(nil)
var _ Product = (*product)(nil)

// ProductRecord — описание продукта на диске (JSON).
type ProductRecord struct {
	ShortNumber     string                       `json:"shortNumber"`
	ExternalID      string                       `json:"externalId"`
	DefaultLanguage string                       `json:"defaultLanguage"`
	Notifications   map[string]map[string]string `json:"notifications"`
}
