package viva_api

import (
	"text/template"

	"github.com/Armor-ru/sds-go/pkg/tplext"
	"github.com/Armor-ru/sds-go/pkg/types"

	"github.com/Armor-ru/viva-api/internal/app/utils"
)

type Viva struct {
	IntTransport types.Transport
	ExtTransport types.Transport

	Secrets []string
	Smpp    utils.SmppConfig

	TestTariffs []string
	Channels    Channels

	SmsTpl     *template.Template
	SmppSender *utils.SmppSender

	VivaPartner PartnerSubscriptionAPI

	AccountId string
}

func (s *Viva) InitSMSInfrastructure() {
	tplText := s.Smpp.Template
	if tplText == "" {
		tplText = "{{.ProductName}}{{ if .Quantity }} for {{ pluralizeEn .Quantity \"device\" \"devices\" }}{{ end }}\n" +
			"Activation code: {{.ActivationCode}}\n" +
			"Download link: {{.DownloadURL}}"
	}
	SmsTpl, _ := template.New("sms").Funcs(tplext.Funcs).Parse(tplText)
	s.SmsTpl = SmsTpl

	s.SmppSender = utils.NewSmppSender(s.Smpp)
}
