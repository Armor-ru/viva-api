package viva_api

import (
	"strings"

	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
)

type inboundSMSDTO struct {
	Phone       string `json:"sourceAddr" validate:"required"`
	ShortNumber string `json:"destinationAddr" validate:"required"`
	Text        string `json:"shortMessage" validate:"required"`
}

func (s *Viva) handleInboundSMS(ctx types.HandlerContext) {
	sms := inboundSMSDTO{}
	ctx.Data(&sms)

	logger.Info().
		Str("phone", sms.Phone).
		Str("shortNumber", sms.ShortNumber).
		Str("text", sms.Text).
		Msg("incoming smpp message")

	s.handleMO(sms)
}

func (s *Viva) handleMO(sms inboundSMSDTO) {
	if s.catalog == nil {
		logger.Error().Msg("incoming smpp ignored, catalog is not loaded")
		return
	}

	productCode, shortNumber, ok := s.catalog.ProductByShortNumber(sms.ShortNumber)
	if !ok {
		logger.Info().Str("shortNumber", sms.ShortNumber).Msg("catalog: unknown short number")
		return
	}

	scenarioKey := moScenarioKey(sms.Text)
	if scenarioKey == "" {
		logger.Info().
			Str("shortNumber", sms.ShortNumber).
			Str("text", sms.Text).
			Msg("catalog: unknown MO command")
		return
	}

	phone := normalizeAddr(sms.Phone)
	if phone == "" {
		logger.Error().Msg("incoming smpp ignored, phone is empty")
		return
	}

	lang := s.catalog.ResolveLang(productCode, moCommandLang(sms.Text))
	ctx := pipeline.Context{
		Phone:       phone,
		ShortNumber: shortNumber,
		ProductCode: productCode,
		Lang:        lang,
	}

	if err := s.runScenario(scenarioKey, ctx); err != nil {
		logger.Error().Err(err).
			Str("phone", phone).
			Str("scenario", scenarioKey).
			Msg("MO scenario failed")
	}
}

func moCommandLang(text string) string {
	switch strings.ToUpper(strings.TrimSpace(text)) {
	case "RUS":
		return "ru"
	case "ARM":
		return "arm"
	case "ENG":
		return "en"
	default:
		return ""
	}
}

func moScenarioKey(text string) string {
	switch strings.ToUpper(strings.TrimSpace(text)) {
	case "1":
		return "mo.1"
	case "STOP":
		return "mo.STOP"
	case "RUS":
		return "mo.RUS"
	case "ARM":
		return "mo.ARM"
	case "ENG":
		return "mo.ENG"
	case "HELP", "?":
		return "mo.help"
	default:
		return ""
	}
}
