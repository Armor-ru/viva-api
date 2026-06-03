package viva_api

import (
	"context"
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
)

// vivaProductName — параметр productName для Viva Init/Confirm.
func vivaProductName(externalID string) string {
	return strings.TrimSpace(externalID)
}

// vivaInitSubscription вызывает Viva InitSubscription (OTP шлёт Viva; подтверждение на лендинге).
func (s *viva) vivaInitSubscription(phone, productName string) error {
	if s.vivaClient == nil {
		return fmt.Errorf("viva api not configured")
	}
	phone = normalizePhone(phone)
	productName = strings.TrimSpace(productName)
	if phone == "" || productName == "" {
		return fmt.Errorf("phone and productName are required")
	}

	c, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	flowInfo(FlowActivation, "6").Str("phone", phone).Str("productName", productName).Msg("calling Viva InitSubscription")

	initRes, err := s.vivaClient.InitSubscription(c, phone, productName)
	if err != nil {
		flowError(FlowActivation, "6").Err(err).Str("phone", phone).Msg("InitSubscription failed")
		return err
	}
	if initRes != nil && initRes.ResultCode != 0 {
		err := fmt.Errorf("%s", vivaSubErr(initRes))
		flowError(FlowActivation, "6").Err(err).Str("phone", phone).Msg("InitSubscription rejected")
		return err
	}

	flowInfo(FlowActivation, "6").Str("phone", phone).Msg("Viva InitSubscription completed")
	return nil
}

// removeVivaSubscription вызывает Viva POST /api/Subscription/RemoveSubscription (сценарий 5).
func (s *viva) removeVivaSubscription(phone, productCode string) error {
	if s.vivaClient == nil {
		return fmt.Errorf("viva api not configured")
	}
	phone = normalizePhone(phone)
	productName := vivaProductName(productCode)
	if phone == "" || productName == "" {
		return fmt.Errorf("phone and productName are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), vivaRequestTimeout)
	defer cancel()

	logger.Info().
		Str("phone", phone).
		Str("productName", productName).
		Msg("scenario 5 step 2: calling Viva /api/Subscription/RemoveSubscription")

	res, err := s.vivaClient.RemoveSubscription(ctx, phone, productName)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("scenario 5 step 2: RemoveSubscription failed")
		return err
	}
	if res != nil && res.ResultCode != 0 {
		err := fmt.Errorf("%s", vivaSubErr(res))
		logger.Error().Err(err).Str("phone", phone).Msg("scenario 5 step 2: RemoveSubscription rejected")
		return err
	}

	logger.Info().
		Str("phone", phone).
		Str("productName", productName).
		Msg("scenario 5 step 2: Viva RemoveSubscription completed")

	return nil
}
