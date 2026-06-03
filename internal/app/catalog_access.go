package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
	"github.com/spf13/cast"
)

func productCodeFromOrder(order types.OrderResponse, item types.OrderItemResponse) string {
	if order.CustomData != nil {
		if code := strings.TrimSpace(cast.ToString(order.CustomData["productCode"])); code != "" {
			return code
		}
	}
	return strings.TrimSpace(cast.ToString(item.Product.Name))
}

// productForShortNumber загружает продукт из файлового каталога по короткому номеру (например 1020).
func (s *viva) productForShortNumber(flow, step, shortNumber string) (catalog.Product, error) {
	if s.catalog == nil {
		flowError(flow, step).Msg("catalog is not loaded")
		return nil, fmt.Errorf("catalog is not loaded")
	}
	shortNumber = strings.TrimSpace(shortNumber)
	if shortNumber == "" {
		return nil, fmt.Errorf("short number is required")
	}

	product, err := s.catalog.GetProductByShortNumber(shortNumber)
	if err != nil {
		flowInfo(flow, step).Err(err).Str("shortNumber", shortNumber).Msg("product not found in file catalog")
		return nil, err
	}

	flowInfo(flow, step).
		Str("shortNumber", product.GetShortNumber()).
		Str("externalId", product.GetExternalID()).
		Msg("product found in file catalog")
	return product, nil
}

// productForExternalID загружает продукт из файлового каталога по externalId (например SAFEKID).
func (s *viva) productForExternalID(flow, step, externalID string) (catalog.Product, error) {
	if s.catalog == nil {
		flowError(flow, step).Msg("catalog is not loaded")
		return nil, fmt.Errorf("catalog is not loaded")
	}
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		return nil, fmt.Errorf("externalId is required")
	}

	product, err := s.catalog.GetProductByExternalId(externalID)
	if err != nil {
		flowInfo(flow, step).Err(err).Str("externalId", externalID).Msg("product not found in file catalog")
		return nil, err
	}

	flowInfo(flow, step).
		Str("externalId", product.GetExternalID()).
		Str("shortNumber", product.GetShortNumber()).
		Msg("product found in file catalog")
	return product, nil
}

// productFromCatalogByExternalID — для цепочек order/completed и webhook.
func (s *viva) productFromCatalogByExternalID(externalID string) (catalog.Product, error) {
	return s.productForExternalID(FlowOrder, "catalog", externalID)
}

func (s *viva) defaultLanguageForProductCode(productCode string) string {
	if s.catalog == nil {
		return "ru"
	}
	product, err := s.catalog.GetProductByExternalId(strings.TrimSpace(productCode))
	if err != nil {
		return "ru"
	}
	return product.GetDefaultLanguage()
}
