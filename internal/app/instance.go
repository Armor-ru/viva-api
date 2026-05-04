package viva_api

import (
	"text/template"

	"github.com/Armor-ru/sds-go/pkg/tplext"
	"github.com/Armor-ru/sds-go/pkg/types"
)

type Viva struct {
	intTransport types.Transport
	extTransport types.Transport

	secrets []string
	smpp    SmppConfig

	testTariffs []string
	channels    Channels

	SmsTpl     *template.Template
	smppSender *SmppSender

	vivaPartner        PartnerSubscriptionAPI
	defaultProductName string
	// orderProductCode — externalId для order/create (тот же UUID, что productCode во вебхуке ActivationRequest).
	orderProductCode string

	accountId string
}

func (s *Viva) InitHandlers() {
	if s.extTransport != nil {
		s.initMiddleWare()

		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivationRequest", s.ExtAppPartnerProductActivationRequestHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductActivation", s.ExtAppPartnerProductActivationHandler)
		s.extTransport.Subscribe("POST /ExtAppPartneerProductRemove", s.ExtAppPartnerProductRemoveHandler)

		s.extTransport.Subscribe("POST /landing/init-subscription", s.LandingInitSubscriptionHandler)
		s.extTransport.Subscribe("POST /landing/confirm-subscription", s.LandingConfirmSubscriptionHandler)
		s.extTransport.Subscribe("GET /landing/subscriber-info/:phoneNum", s.LandingGetSubscriberInfoGETHandler)
		s.extTransport.Subscribe("POST /landing/subscriber-info", s.LandingGetSubscriberInfoPOSTHandler)

		s.extTransport.Subscribe("POST /landing/:locale/init-subscription", s.LandingInitSubscriptionLocalizedHandler)
		s.extTransport.Subscribe("POST /landing/:locale/confirm-subscription", s.LandingConfirmSubscriptionLocalizedHandler)
		s.extTransport.Subscribe("GET /landing/:locale/subscriber-info/:phoneNum", s.LandingGetSubscriberInfoGETLocalizedHandler)
		s.extTransport.Subscribe("POST /landing/:locale/subscriber-info", s.LandingGetSubscriberInfoPOSTLocalizedHandler)
	}

	if s.intTransport != nil {
		s.intTransport.Subscribe("order/completed", s.onCompletedHandler)
	}

	tplText := s.smpp.Template
	if tplText == "" {
		tplText = "{{.ProductName}}{{ if .Quantity }} for {{ pluralizeEn .Quantity \"device\" \"devices\" }}{{ end }}\n" +
			"Activation code: {{.ActivationCode}}\n" +
			"Download link: {{.DownloadURL}}"
	}
	SmsTpl, _ := template.New("sms").Funcs(tplext.Funcs).Parse(tplText)
	s.SmsTpl = SmsTpl

	s.smppSender = NewSmppSender(s.smpp)
}
