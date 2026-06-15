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
	"git.dev.armlab.pro/armor/sds-go/pkg/types"

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

	SMPP struct {
		Endpoint []string `yaml:"endpoint"`
		Auth     struct {
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		} `yaml:"auth"`
		Address struct {
			SourceAddr    string `yaml:"sourceAddr"`
			SourceAddrTON uint8  `yaml:"sourceAddrTON"`
			SourceAddrNPI uint8  `yaml:"sourceAddrNPI"`
			DestAddrTON   uint8  `yaml:"destAddrTON"`
			DestAddrNPI   uint8  `yaml:"destAddrNPI"`
		} `yaml:"address"`
		RespTimeout time.Duration `yaml:"respTimeout"`
	} `yaml:"smpp"`
	CatalogDir        string `yaml:"catalogDir"`
	AccountId         string `yaml:"accountId"`
	LandingConfirmURL string `yaml:"landingConfirmURL"`

	VivaAPI struct {
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

	smppRespTimeout := cfg.SMPP.RespTimeout
	if smppRespTimeout <= 0 {
		smppRespTimeout = 10 * time.Second
	}

	smpp := transportSmpp.NewTransport(transportSmpp.Config{
		Name:      ServiceName,
		Endpoints: cfg.SMPP.Endpoint,
		User:      cfg.SMPP.Auth.User,
		Passwd:    cfg.SMPP.Auth.Password,
		Address: transportSmpp.AddressConfig{
			SourceAddr:    cfg.SMPP.Address.SourceAddr,
			SourceAddrTON: cfg.SMPP.Address.SourceAddrTON,
			SourceAddrNPI: cfg.SMPP.Address.SourceAddrNPI,
			DestAddrTON:   cfg.SMPP.Address.DestAddrTON,
			DestAddrNPI:   cfg.SMPP.Address.DestAddrNPI,
		},
		RespTimeout: smppRespTimeout,
	})
	defer smpp.Disconnect()

	if tr, ok := smpp.(*transportSmpp.Transport); ok {
		tr.Error(func(err error, _ types.HandlerContext) {
			logger.Error().Msg("smpp inbound handler failed, " + err.Error())
		})
	}

	var client *vivaclient.Client
	if strings.TrimSpace(cfg.VivaAPI.BaseURL) != "" {
		client = vivaclient.New(vivaclient.Config{
			BaseURL:  cfg.VivaAPI.BaseURL,
			UserName: cfg.VivaAPI.UserName,
			Password: cfg.VivaAPI.Password,
		})
	}

	viva_api.New(
		viva_api.WithIntTransport(nats),
		viva_api.WithExtTransport(http),
		viva_api.WithUssdTransport(smpp),
		viva_api.WithSecrets(cfg.ExtTransport.Secret),
		viva_api.WithCatalogDir(cfg.CatalogDir),
		viva_api.WithAccountId(cfg.AccountId),
		viva_api.WithVivaClient(client),
		viva_api.WithLandingConfirmURL(cfg.LandingConfirmURL),
	)
	if err := http.Connect(); err != nil {
		logger.Error().Msg("http connect failed, " + err.Error())
	}

	if err := smpp.Connect(); err != nil {
		logger.Error().Msg("smpp connect failed, " + err.Error())
	}

	if err := nats.ConnectAndWait(); err != nil {
		logger.Fatal().Msg("nats connect failed, " + err.Error())
	}
}
