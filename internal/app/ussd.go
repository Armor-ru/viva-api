package viva_api

import (
	"git.dev.armlab.pro/armor/sds-go/pkg/types"
)

type inboundSMSDTO struct {
	Phone       smppField `json:"sourceAddr"`
	ShortNumber smppField `json:"destinationAddr"`
	Text        string    `json:"shortMessage"`
}

// ussdHandler — входящие MO по ussdTransport (топик smpp/inbound), SPEC: ussdHandler.
func (s *viva) ussdHandler(ctx types.HandlerContext) {
	var sms inboundSMSDTO
	ctx.Data(&sms)
	mo := parseInboundMO(sms)

	if mo.isActivationOne() {
		s.handleActivationOne(mo)
		return
	}
	if mo.isStopOn1020() {
		s.handleStopOn1020(mo)
		return
	}
	if mo.isLanguageChangeOn1020() {
		s.handleLanguageOn1020(mo)
		return
	}
	if mo.isUnknownCommandOn1020() {
		s.handleUnknownOn1020(mo)
		return
	}

	flowInfo(FlowUnknown, "0").
		Str("phone", mo.Phone).
		Str("shortNumber", mo.ShortNumber).
		Str("text", mo.Text).
		Msg("unrouted MO — no handler for this short number")
}
