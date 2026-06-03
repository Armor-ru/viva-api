package viva_api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.dev.armlab.pro/armor/viva-api/internal/vivaclient"
)

// vivaAPITestDouble — счётчики вызовов Viva API для unit-тестов.
type vivaAPITestDouble struct {
	InitCalls    int
	ConfirmCalls int
	RemoveCalls  int
	ConfirmOTP   string
	LastRemove   struct {
		Phone, Product string
	}
	InitErr    error
	ConfirmErr error
	RemoveErr  error

	srv *httptest.Server
}

func newVivaAPITestDouble(t *testing.T) *vivaAPITestDouble {
	t.Helper()
	d := &vivaAPITestDouble{}
	d.srv = httptest.NewServer(http.HandlerFunc(d.serve))
	t.Cleanup(d.srv.Close)
	return d
}

func (d *vivaAPITestDouble) Client() *vivaclient.Client {
	return vivaclient.NewWithHTTP(d.srv.Client(), d.srv.URL, "test", "test")
}

func (d *vivaAPITestDouble) serve(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/auth/token":
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"expires_in":   3600,
		})
	case strings.Contains(path, "InitSubscription"):
		d.InitCalls++
		if d.InitErr != nil {
			http.Error(w, d.InitErr.Error(), http.StatusBadGateway)
			return
		}
		writeVivaOK(w)
	case strings.Contains(path, "ConfirmSubscription"):
		d.ConfirmCalls++
		d.ConfirmOTP = strings.TrimSpace(r.URL.Query().Get("otp"))
		if d.ConfirmErr != nil {
			http.Error(w, d.ConfirmErr.Error(), http.StatusBadGateway)
			return
		}
		writeVivaOK(w)
	case strings.Contains(path, "RemoveSubscription"):
		d.RemoveCalls++
		d.LastRemove.Phone = strings.TrimSpace(r.URL.Query().Get("phoneNum"))
		d.LastRemove.Product = strings.TrimSpace(r.URL.Query().Get("productName"))
		if d.RemoveErr != nil {
			http.Error(w, d.RemoveErr.Error(), http.StatusBadGateway)
			return
		}
		writeVivaOK(w)
	default:
		http.NotFound(w, r)
	}
}

func writeVivaOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(vivaclient.ResponseModel{ResultCode: 0, Result: true})
}
