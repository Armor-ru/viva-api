package viva_api

import (
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
)

// Имена потоков для структурированных логов (flow + step).
const (
	FlowActivation   = "activation"      // сценарии 1–2, MO «1»
	FlowStop         = "stop"            // сценарии 5–6
	FlowLang         = "lang"            // сценарий 7
	FlowUnknown      = "unknown"         // сценарий 8
	FlowTrial        = "trial_reminder"  // сценарий 3
	FlowPaid         = "paid_activation" // сценарий 4
	FlowOrder        = "order"           // NATS: order/create, order/completed
)

func flowInfo(flow, step string) *logger.Event {
	return logger.Info().Str("flow", flow).Str("step", step)
}

func flowError(flow, step string) *logger.Event {
	return logger.Error().Str("flow", flow).Str("step", step)
}
