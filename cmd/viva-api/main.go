package main

import (
	"flag"
	"strings"
	"time"

	"git.dev.armlab.pro/armor/sds-go/pkg/config"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
	"git.dev.armlab.pro/armor/sds-go/pkg/systemd"
	utilsTls "git.dev.armlab.pro/armor/sds-go/pkg/tls"
	"git.dev.armlab.pro/armor/sds-go/pkg/transport"
	transportSmpp "git.dev.armlab.pro/armor/sds-go/pkg/transport/smpp"

	viva_api "git.dev.armlab.pro/armor/viva-api/internal/app"
	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
)

var ServiceName = "viva-api"
var ConfFile = "./config/viva-api.yaml"

type Conf struct {
	IntTransport transport.ConfigTransport `yaml:"intTransport"`
	ExtTransport struct {
		Server string   `yaml:"server"`
		Secret []string `yaml:"secret"`
	} `yaml:"extTransport"`
	SMPP      viva_api.SmppConfig `yaml:"smpp"`
	App       viva_api.AppConfig  `yaml:"app"`
	AccountId string              `yaml:"accountId"`
	VivaAPI   struct {
		BaseURL  string `yaml:"baseURL"`
		UserName string `yaml:"userName"`
		Password string `yaml:"password"`
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

	notify := transportSmpp.NewTransport(transportSmpp.Config{
		Name:        ServiceName,
		Endpoints:   cfg.SMPP.Endpoint,
		User:        cfg.SMPP.Auth.User,
		Passwd:        cfg.SMPP.Auth.Password,
		RespTimeout:   5 * time.Second,
		Address: transportSmpp.AddressConfig{
			SourceAddr:    cfg.SMPP.Address.SourceAddr,
			SourceAddrTON: cfg.SMPP.Address.SourceAddrTON,
			SourceAddrNPI: cfg.SMPP.Address.SourceAddrNPI,
			DestAddrTON:   cfg.SMPP.Address.DestAddrTON,
			DestAddrNPI:   cfg.SMPP.Address.DestAddrNPI,
		},
	})
	defer notify.Disconnect()

	var client *vivaclient.Client
	if strings.TrimSpace(cfg.VivaAPI.BaseURL) != "" {
		client = vivaclient.New(vivaclient.Config{
			BaseURL:  cfg.VivaAPI.BaseURL,
			UserName: cfg.VivaAPI.UserName,
			Password: cfg.VivaAPI.Password,
		})
	}

	catalogDir := strings.TrimSpace(cfg.App.CatalogDir)
	if catalogDir == "" {
		catalogDir = "catalog"
	}

	viva_api.New(
		viva_api.WithIntTransport(nats),
		viva_api.WithExtTransport(http),
		viva_api.WithUssdTransport(notify),
		viva_api.WithSecrets(cfg.ExtTransport.Secret),
		viva_api.WithCatalogDir(catalogDir),
		viva_api.WithDefaultLanguage(cfg.App.ResolvedDefaultLanguage(), cfg.App.ResolvedLangPreferenceTTL()),
		viva_api.WithAccountId(cfg.AccountId),
		viva_api.WithVivaClient(client),
		viva_api.WithLandingConfirmURL(cfg.App.LandingConfirmURL),
	)

	nats.ConnectAndWait()
}
