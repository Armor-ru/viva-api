package viva

import (
	"os"
	"testing"
	"time"

	viva_api "github.com/Armor-ru/viva-api/internal/app"
	"github.com/fiorix/go-smpp/smpp"
)

// Это интеграционные тесты для проверки отправки и подключения к smpp

func TestSMPP_Send_Integration(t *testing.T) {
	if os.Getenv("SMPP_IT") != "1" {
		t.Skip("skip smpp integration test; set SMPP_IT=1 to run")
	}

	endpoint := os.Getenv("SMPP_ENDPOINT")
	user := os.Getenv("SMPP_USER")
	pass := os.Getenv("SMPP_PASS")
	msisdn := os.Getenv("SMPP_MSISDN")
	text := os.Getenv("SMPP_TEXT")

	if endpoint == "" || user == "" || pass == "" || msisdn == "" {
		t.Fatalf("missing env: SMPP_ENDPOINT, SMPP_USER, SMPP_PASS, SMPP_MSISDN (optional SMPP_TEXT)")
	}
	if text == "" {
		text = "viva-api smpp integration test"
	}

	cfg := viva_api.SmppConfig{Endpoint: []string{endpoint}}
	cfg.Auth.User = user
	cfg.Auth.Password = pass

	sender := viva_api.NewSmppSender(cfg)

	if err := sender.Send(msisdn, text); err != nil {
		t.Fatalf("smpp send failed: %v", err)
	}
}

func TestSMPP_Bind_Integration(t *testing.T) {
	if os.Getenv("SMPP_IT") != "1" {
		t.Skip("skip smpp integration test; set SMPP_IT=1 to run")
	}

	endpoint := os.Getenv("SMPP_ENDPOINT")
	user := os.Getenv("SMPP_USER")
	pass := os.Getenv("SMPP_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Fatalf("missing env: SMPP_ENDPOINT, SMPP_USER, SMPP_PASS")
	}

	tx := &smpp.Transceiver{
		Addr:   endpoint,
		User:   user,
		Passwd: pass,
	}

	conn := tx.Bind()
	defer tx.Close()

	select {
	case st := <-conn:
		if st.Error() != nil {
			t.Fatalf("bind failed: %v", st.Error())
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("bind timeout")
	}
}
