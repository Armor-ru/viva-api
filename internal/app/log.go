package viva_api

import (
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
)

func logAppError(err error, msg string) {
	if err == nil {
		return
	}
	logger.Error().Fields(errs.Fields(err)).Msg(msg)
}

func logUssdInitDone(phone, product string, resultCode int, next string) {
	logger.Info().
		Str("phone", phone).
		Str("product", product).
		Int("resultCode", resultCode).
		Str("next", next).
		Msg("ussd init done")
}

func logUssdStopDone(phone, product, next string) {
	logger.Info().
		Str("phone", phone).
		Str("product", product).
		Str("next", next).
		Msg("ussd stop done")
}

func logUssdLangDone(phone, product, lang string) {
	logger.Info().
		Str("phone", phone).
		Str("product", product).
		Str("lang", lang).
		Msg("ussd lang done")
}

func logUssdUnknownDone(phone, product, text, lang string) {
	logger.Info().
		Str("phone", phone).
		Str("product", product).
		Str("text", text).
		Str("lang", lang).
		Msg("ussd unknown done")
}

func logLandingConfirmDone(phone, orderID string) {
	logger.Info().
		Str("phone", phone).
		Str("orderId", orderID).
		Msg("landing confirm done")
}

func logOrderExpiresDone(orderID, phone, product string, endTime time.Time) {
	logger.Info().
		Str("orderId", orderID).
		Str("phone", phone).
		Str("product", product).
		Time("endTime", endTime).
		Str("notify", "trial_expires").
		Msg("order expires done")
}

func logWebhookRenewDone(phone, product, orderID string) {
	logger.Info().
		Str("phone", phone).
		Str("product", product).
		Str("orderId", orderID).
		Msg("webhook renew done")
}

func logWebhookRemoveDone(phone, product, orderID string) {
	logger.Info().
		Str("phone", phone).
		Str("product", product).
		Str("orderId", orderID).
		Str("notify", "service_deactivated").
		Msg("webhook remove done")
}

func logOrderCompleteDone(orderID, phone, product, itemType, welcomeNotify string) {
	logger.Info().
		Str("orderId", orderID).
		Str("phone", phone).
		Str("product", product).
		Str("itemType", itemType).
		Str("notify", welcomeNotify).
		Str("alsoNotify", "license").
		Msg("order complete done")
}
