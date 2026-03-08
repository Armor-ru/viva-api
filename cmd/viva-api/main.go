package main

import (
	"flag"

	"github.com/Armor-ru/sds-go/pkg/config"
	"github.com/Armor-ru/sds-go/pkg/logger"
	"github.com/Armor-ru/sds-go/pkg/systemd"
	utilsTls "github.com/Armor-ru/sds-go/pkg/tls"
	"github.com/Armor-ru/sds-go/pkg/transport"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
)

var ServiceName = "viva-api"
var ConfFile = "./config/viva-api.yaml"

type Conf struct {
	IntTransport transport.ConfigTransport `yaml:"intTransport"`
	ExtTransport struct {
		Server string   `yaml:"server"`
		Secret []string `yaml:"secret"`
	} `yaml:"extTransport"`

	SMPP        viva_api.SmppConfig `yaml:"smpp"`
	TestTariffs []string            `yaml:"testTariffs"`
	Channels    viva_api.Channels   `yaml:"channels"`
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

	viva_api.New(
		viva_api.WithIntTransport(nats),
		viva_api.WithExtTransport(http),
		viva_api.WithSecrets(cfg.ExtTransport.Secret),
		viva_api.WithSmppConfig(cfg.SMPP),
	)

	nats.ConnectAndWait()
}
