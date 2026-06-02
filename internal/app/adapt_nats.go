package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
	"github.com/spf13/cast"
)

func (s *Viva) onCompletedHandler(ctx types.HandlerContext) {
	order := types.OrderResponse{}
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.completed")

	if err := s.handleOrderCompleted(order); err != nil {
		logger.Error().Err(err).Str("orderId", order.ID).Msg("order.completed pipeline failed")
	}
}

func (s *Viva) onExpiresHandler(ctx types.HandlerContext) {
	order := types.OrderResponse{}
	ctx.Data(&order)

	logger.Info().Interface("order", order).Msg("receive order.expires")

	if err := s.handleOrderExpires(order); err != nil {
		logger.Error().Err(err).Str("orderId", order.ID).Msg("order.expires pipeline failed")
	}
}

func (s *Viva) handleOrderCompleted(order types.OrderResponse) error {
	if order.Status == "error" {
		logger.Info().Str("orderId", order.ID).Msg("order status error, skip notify")
		return nil
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order %s: items is empty", order.ID)
	}

	hasActivation := false
	for _, it := range order.Items {
		if it.Type == "activate" || it.Type == "reactivate" {
			hasActivation = true
			break
		}
	}
	if !hasActivation {
		logger.Info().Str("orderId", order.ID).Msg("no activation item, skip notify")
		return nil
	}

	ctx, err := s.pipelineContextFromOrder(order)
	if err != nil {
		return err
	}
	return s.runScenario("nats.order.completed", ctx)
}

func (s *Viva) handleOrderExpires(order types.OrderResponse) error {
	ctx, err := s.pipelineContextBase(order)
	if err != nil {
		return err
	}
	return s.runScenario("nats.order.expires", ctx)
}

func (s *Viva) pipelineContextFromOrder(order types.OrderResponse) (pipeline.Context, error) {
	ctx, err := s.pipelineContextBase(order)
	if err != nil {
		return pipeline.Context{}, err
	}
	if len(order.Items) == 0 {
		return pipeline.Context{}, fmt.Errorf("order %s: items is empty", order.ID)
	}

	item := order.Items[0]
	ctx.ProductName = cast.ToString(item.Product.Name)
	ctx.ActivationCode = strings.TrimSpace(cast.ToString(item.Artifacts["ActivationCode"]))
	ctx.DownloadURL = firstDownloadURL(item.Artifacts["download"])
	if quantity, ok := item.Options["quantity"]; ok {
		ctx.Quantity = cast.ToInt(quantity)
	}
	ctx.Order = &order
	return ctx, nil
}

func (s *Viva) pipelineContextBase(order types.OrderResponse) (pipeline.Context, error) {
	if s.catalog == nil {
		return pipeline.Context{}, fmt.Errorf("catalog is not loaded")
	}

	productCode := extractProductCode(order)
	if productCode == "" {
		return pipeline.Context{}, fmt.Errorf("order %s: product external id is empty", order.ID)
	}

	_, shortNumber, _ := s.catalog.ProductByExternalID(productCode)

	phone := normalizeAddr(cast.ToString(order.Fields["phone"]))
	if phone == "" {
		return pipeline.Context{}, fmt.Errorf("order %s: phone is empty", order.ID)
	}

	return pipeline.Context{
		Phone:       phone,
		ShortNumber: shortNumber,
		ProductCode: productCode,
		Lang:        orderLang(order, s.catalog, productCode),
	}, nil
}

func extractProductCode(order types.OrderResponse) string {
	if len(order.Items) > 0 && len(order.Items[0].Tariff.ExternalID) > 0 {
		if code := strings.TrimSpace(order.Items[0].Tariff.ExternalID[0]); code != "" {
			return code
		}
	}
	if order.CustomData != nil {
		if code := strings.TrimSpace(cast.ToString(order.CustomData["productCode"])); code != "" {
			return code
		}
	}
	return ""
}

func firstDownloadURL(raw interface{}) string {
	list, ok := raw.([]interface{})
	if !ok {
		return ""
	}
	for _, v := range list {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if u := strings.TrimSpace(cast.ToString(m["url"])); u != "" {
			return u
		}
	}
	return ""
}
