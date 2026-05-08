package vivaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const renewSkew = 5 * time.Minute

type Config struct {
	BaseURL  string
	UserName string
	Password string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type Client struct {
	http    *http.Client
	baseURL string
	auth    *tokenManager
}

func New(cfg Config) *Client {
	return &Client{
		http:    http.DefaultClient,
		baseURL: normalizeBaseURL(cfg.BaseURL),
		auth:    newTokenManager(cfg.UserName, cfg.Password),
	}
}

func normalizeBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

type tokenManager struct {
	mu       sync.Mutex
	userName string
	password string
	token    string
	expiry   time.Time
}

func newTokenManager(userName, password string) *tokenManager {
	return &tokenManager{
		userName: userName,
		password: password,
	}
}

func (tm *tokenManager) get(ctx context.Context, client *http.Client, baseURL string) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.isValid() {
		return tm.token, nil
	}

	token, err := tm.fetch(ctx, client, baseURL)
	if err != nil {
		return "", fmt.Errorf("fetch token: %w", err)
	}

	tm.token = token.AccessToken
	tm.expiry = time.Now().Add(token.expiryDuration())
	return tm.token, nil
}

func (tm *tokenManager) isValid() bool {
	return tm.token != "" && time.Until(tm.expiry) > renewSkew
}

func (tm *tokenManager) fetch(ctx context.Context, client *http.Client, baseURL string) (*TokenResponse, error) {
	body, err := json.Marshal(map[string]string{
		"userName": tm.userName,
		"password": tm.password,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal auth body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL+"/auth/token",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute auth request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("viva auth: status %d: %s", resp.StatusCode, string(raw))
	}

	var tr TokenResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, fmt.Errorf("unmarshal token response: %w", err)
	}

	if tr.AccessToken == "" {
		return nil, fmt.Errorf("viva auth: empty access_token")
	}

	return &tr, nil
}

func (tr TokenResponse) expiryDuration() time.Duration {
	if tr.ExpiresIn <= 0 {
		return time.Hour // дефолтное значение
	}
	return time.Duration(tr.ExpiresIn) * time.Second
}

// --- HTTP-методы клиента ---

// do выполняет HTTP-запрос с авторизацией и парсингом ответа
func (c *Client) do(ctx context.Context, method, path string, body io.Reader, out interface{}) error {
	token, err := c.auth.get(ctx, c.http, c.baseURL)
	if err != nil {
		return fmt.Errorf("get auth token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{
			StatusCode: resp.StatusCode,
			Message:    string(raw),
			Path:       path,
		}
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) GetSubscriberInfo(ctx context.Context, msisdn string) (*GetSubInfoResponse, error) {
	path := fmt.Sprintf("/api/Subscriber/%s", strings.TrimSpace(msisdn))

	var out GetSubInfoResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, fmt.Errorf("get subscriber info: %w", err)
	}

	return &out, nil
}

func (c *Client) InitSubscription(ctx context.Context, phoneNum, productName string) (*ResponseModel, error) {
	path := subscriptionPath("/api/Subscription/InitSubscription", phoneNum, productName, nil)

	var out ResponseModel
	if err := c.do(ctx, http.MethodPost, path, nil, &out); err != nil {
		return nil, fmt.Errorf("init subscription: %w", err)
	}

	return &out, nil
}

func (c *Client) ConfirmSubscription(ctx context.Context, phoneNum, productName string, otp *string) (*ResponseModel, error) {
	path := subscriptionPath("/api/Subscription/ConfirmSubscription", phoneNum, productName, otp)

	var out ResponseModel
	if err := c.do(ctx, http.MethodPost, path, nil, &out); err != nil {
		return nil, fmt.Errorf("confirm subscription: %w", err)
	}

	return &out, nil
}

func subscriptionPath(basePath, phoneNum, productName string, otp *string) string {
	q := url.Values{}
	q.Set("phoneNum", strings.TrimSpace(phoneNum))
	q.Set("productName", strings.TrimSpace(productName))
	if otp != nil {
		q.Set("otp", strings.TrimSpace(*otp))
	}
	return basePath + "?" + q.Encode()
}

type Error struct {
	StatusCode int
	Message    string
	Path       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("viva %s: status %d: %s", e.Path, e.StatusCode, e.Message)
}

func (e *Error) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
