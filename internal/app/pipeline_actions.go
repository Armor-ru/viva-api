package viva_api

import (
	"fmt"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
	"github.com/google/uuid"
	"github.com/spf13/cast"
)

type vivaActions struct {
	v *Viva
}

func (a vivaActions) CreateOrder(orderType types.OrderType, ctx pipeline.Context) error {
	_, err := a.v.createOrder(orderType, ctx.Phone, ctx.ProductCode, ctx.Lang)
	return err
}

func (a vivaActions) Notify(ctx pipeline.Context, tplKey string) error {
	if a.v.catalog == nil {
		return fmt.Errorf("catalog is not loaded")
	}
	data := notifyTemplateData(ctx)
	text, err := a.v.catalog.RenderNotify(ctx.ProductCode, tplKey, ctx.Lang, data)
	if err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("empty notification for tpl %q", tplKey)
	}
	logger.Info().Str("phone", ctx.Phone).Str("tpl", tplKey).Str("text", text).Msg("send sms notify")
	return a.v.notify(ctx.Phone, text)
}

func notifyTemplateData(ctx pipeline.Context) map[string]interface{} {
	return map[string]interface{}{
		"Phone":          ctx.Phone,
		"ShortNumber":    ctx.ShortNumber,
		"ExternalID":     ctx.ProductCode,
		"ProductName":    ctx.ProductName,
		"ActivationCode": ctx.ActivationCode,
		"DownloadURL":    ctx.DownloadURL,
		"Quantity":       ctx.Quantity,
		"Language":       ctx.Lang,
	}
}

func (s *Viva) runScenario(scenarioKey string, ctx pipeline.Context) error {
	if s.engine == nil {
		return fmt.Errorf("pipeline engine is not configured")
	}
	return s.engine.Run(scenarioKey, ctx)
}

func (s *Viva) createOrder(orderType types.OrderType, phone, externalID, lang string) (string, error) {
	if s.intTransport == nil {
		return "", fmt.Errorf("intTransport is not configured")
	}

	phone = normalizeAddr(phone)
	externalID = strings.TrimSpace(externalID)
	if phone == "" || externalID == "" {
		return "", fmt.Errorf("phoneNum and productCode are required")
	}

	orderId := uuid.NewSHA1(uuid.MustParse(s.accountId), []byte(externalID+":"+phone)).String()

	var items []types.OrderItemRequest
	if orderType == types.OrderTypeNew {
		items = append(items, types.OrderItemRequest{
			ID:         uuid.NewString(),
			ExternalID: &externalID,
		})
	}

	req := types.OrderCreateRequest{
		ID:   orderId,
		Type: orderType,
		Fields: types.JSON{
			"phone": phone,
		},
		Items: items,
	}
	if l := normalizeLanguage(lang); l != "" {
		req.CustomData = map[string]interface{}{"lang": l}
	}

	_, err := s.intTransport.Send("order/create", req, types.SendOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("send order/create: %w", err)
	}
	return orderId, nil
}

func normalizeAddr(phone string) string {
	return strings.TrimPrefix(strings.TrimSpace(phone), "+")
}

func normalizeLanguage(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "arm", "hy":
		return "arm"
	case "rus", "ru":
		return "ru"
	case "eng", "en":
		return "en"
	default:
		return ""
	}
}

func orderLang(order types.OrderResponse, catalog pipeline.Catalog, productCode string) string {
	if order.CustomData != nil {
		if l := normalizeLanguage(cast.ToString(order.CustomData["lang"])); l != "" {
			return l
		}
	}
	if catalog != nil {
		return catalog.ResolveLang(productCode, "")
	}
	return ""
}
