package pipeline

import "git.dev.armlab.pro/armor/sds-go/pkg/types"

// Context carries data for any scenario (MO, NATS, webhook).
type Context struct {
	Phone       string
	ShortNumber string
	ProductCode string
	Lang        string

	ActivationCode string
	DownloadURL    string
	ProductName    string
	Quantity       int

	Order *types.OrderResponse
}

// Actions are side effects invoked by pipeline steps.
type Actions interface {
	CreateOrder(orderType types.OrderType, ctx Context) error
	Notify(ctx Context, tplKey string) error
}

// Catalog supplies scenario steps and notification templates.
type Catalog interface {
	Steps(productCode, scenarioKey string) ([]Step, error)
	RenderNotify(productCode, tplKey, lang string, data map[string]interface{}) (string, error)
	ResolveLang(productCode, langHint string) string
	ProductByShortNumber(short string) (productCode, shortNumber string, ok bool)
	ProductByExternalID(externalID string) (productCode, shortNumber string, ok bool)
}

type Step struct {
	Command string
	Params  map[string]interface{}
}
