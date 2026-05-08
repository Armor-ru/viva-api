package main

import (
	"flag"
	"strings"

	"github.com/Armor-ru/sds-go/pkg/config"
	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/systemd"
	utilsTls "github.com/Armor-ru/sds-go/pkg/tls"
	"github.com/Armor-ru/sds-go/pkg/transport"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
	"github.com/Armor-ru/viva-api/internal/app/handlers"
	"github.com/Armor-ru/viva-api/internal/app/utils"
	"github.com/Armor-ru/viva-api/internal/vivaclient"
)

var ServiceName = "viva-api"
var ConfFile = "./config/viva-api.yaml"

type Conf struct {
	IntTransport transport.ConfigTransport `yaml:"intTransport"`
	ExtTransport struct {
		Server string   `yaml:"server"`
		Secret []string `yaml:"secret"`
	} `yaml:"extTransport"`

	SMPP        utils.SmppConfig  `yaml:"smpp"`
	TestTariffs []string          `yaml:"testTariffs"`
	Channels    viva_api.Channels `yaml:"channels"`
	AccountId   string            `yaml:"accountId"`

	VivaAPI struct {
		BaseURL          string `yaml:"baseURL"`
		UserName         string `yaml:"userName"`
		Password         string `yaml:"password"`
		DefaultProduct   string `yaml:"defaultProduct"`
		OrderProductCode string `yaml:"orderProductCode"`
	} `yaml:"vivaApi"`
}

func init() {
	flag.StringVar(&ConfFile, "cf", ConfFile, "config file")
}

func main() {
	flag.Parse()
	command := flag.Arg(0)

	if command != "" {
		systemd.RunCommand(command, ServiceName, ConfFile)
		return
	}

	logger.InitWithContext(ServiceName)

	cfg := Conf{}
	config.Get(ConfFile, &cfg)

	nats := transport.NewNatsTransport(transport.Config{
		Name:   ServiceName,
		Server: cfg.IntTransport.Server,
		Tls:    utilsTls.NewTls(cfg.IntTransport.Tls),
	})
	defer nats.Disconnect()

	http := transport.NewHttpTransport(transport.Config{
		Name:   ServiceName,
		Server: cfg.ExtTransport.Server,
	})
	defer http.Disconnect()
	http.Connect()

	var partner viva_api.PartnerSubscriptionAPI
	if strings.TrimSpace(cfg.VivaAPI.BaseURL) != "" {
		partner = vivaclient.New(vivaclient.Config{
			BaseURL:  strings.TrimSpace(cfg.VivaAPI.BaseURL),
			UserName: strings.TrimSpace(cfg.VivaAPI.UserName),
			Password: strings.TrimSpace(cfg.VivaAPI.Password),
		})
	}

	v := viva_api.New(
		viva_api.WithIntTransport(nats),
		viva_api.WithExtTransport(http),
		viva_api.WithSecrets(cfg.ExtTransport.Secret),
		viva_api.WithSmppConfig(cfg.SMPP),
		viva_api.WithAccountId(cfg.AccountId),
		viva_api.WithVivaPartner(partner),
		viva_api.WithDefaultProductName(cfg.VivaAPI.DefaultProduct),
		viva_api.WithOrderProductCode(cfg.VivaAPI.OrderProductCode),
	)
	handlers.Register(&v)

	nats.ConnectAndWait()
}
