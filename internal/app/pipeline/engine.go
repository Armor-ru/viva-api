package pipeline

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"github.com/spf13/cast"
)

type Engine struct {
	Catalog Catalog
	Actions Actions
}

func (e *Engine) Run(scenarioKey string, ctx Context) error {
	if e == nil || e.Catalog == nil || e.Actions == nil {
		return fmt.Errorf("pipeline engine is not configured")
	}

	productCode := strings.TrimSpace(ctx.ProductCode)
	if productCode == "" {
		return fmt.Errorf("productCode is required")
	}

	steps, err := e.Catalog.Steps(productCode, scenarioKey)
	if err != nil {
		return err
	}

	for _, step := range steps {
		switch normalizeKey(step.Command) {
		case "new":
			if err := e.Actions.CreateOrder(types.OrderTypeNew, ctx); err != nil {
				return fmt.Errorf("new: %w", err)
			}
		case "renew":
			if err := e.Actions.CreateOrder(types.OrderTypeRenew, ctx); err != nil {
				return fmt.Errorf("renew: %w", err)
			}
		case "cancel":
			if err := e.Actions.CreateOrder(types.OrderTypeCancel, ctx); err != nil {
				return fmt.Errorf("cancel: %w", err)
			}
		case "notify":
			tpl := strings.TrimSpace(cast.ToString(step.Params["tpl"]))
			if tpl == "" {
				return fmt.Errorf("notify: params.tpl is required")
			}
			if err := e.Actions.Notify(ctx, tpl); err != nil {
				return fmt.Errorf("notify %q: %w", tpl, err)
			}
		default:
			return fmt.Errorf("unsupported command %q", step.Command)
		}
	}
	return nil
}

func normalizeKey(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
