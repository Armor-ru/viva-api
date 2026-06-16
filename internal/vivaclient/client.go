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

const tokenRenewSkew = 5 * time.Minute

type Config struct {
	BaseURL  string
	UserName string
	Password string
}

type Client struct {
	http    *http.Client
	baseURL string
	auth    *tokenManager
}

func New(cfg Config) *Client {
	return &Client{
		http:    http.DefaultClient,
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		auth:    newTokenManager(cfg.UserName, cfg.Password),
	}
}

func (c *Client) InitSubscription(ctx context.Context, phoneNum, productName string) (*ResponseModel, error) {
	path := subscriptionPath("/api/Subscription/InitSubscription", phoneNum, productName, nil)

	var out ResponseModel
	if err := c.do(ctx, path, &out); err != nil {
		return nil, fmt.Errorf("init subscription failed, %w", err)
	}

	return &out, nil
}

func (c *Client) ConfirmSubscription(ctx context.Context, phoneNum, productName string, otp string) (*ResponseModel, error) {
	path := subscriptionPath("/api/Subscription/ConfirmSubscription", phoneNum, productName, &otp)

	var out ResponseModel
	if err := c.do(ctx, path, &out); err != nil {
		return nil, fmt.Errorf("confirm subscription failed, %w", err)
	}

	return &out, nil
}

func (c *Client) RemoveSubscription(ctx context.Context, phoneNum, productName string) (*ResponseModel, error) {

	path := subscriptionPath("/api/Subscription/RemoveSubscription", phoneNum, productName, nil)
	var out ResponseModel
	if err := c.do(ctx, path, &out); err != nil {
		return nil, fmt.Errorf("remove subscription failed, %w", err)
	}
	return &out, nil

}

func (c *Client) do(ctx context.Context, path string, out interface{}) error {
	token, err := c.auth.get(ctx, c.http, c.baseURL)
	if err != nil {
		return fmt.Errorf("get auth token failed, %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request failed, %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("execute request failed, %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed, %w", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("viva %s, status %d, %s", path, resp.StatusCode, string(raw))
		}
		return fmt.Errorf("unmarshal response failed, %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if rm, ok := out.(*ResponseModel); ok && rm.ResultCode != 0 {
			return nil
		}
		return fmt.Errorf("viva %s, status %d, %s", path, resp.StatusCode, string(raw))
	}
	return nil
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

type tokenManager struct {
	mu       sync.Mutex
	userName string
	password string
	token    string
	expiry   time.Time
}

func newTokenManager(userName, password string) *tokenManager {
	return &tokenManager{
		userName: strings.TrimSpace(userName),
		password: strings.TrimSpace(password),
	}
}

func (tm *tokenManager) get(ctx context.Context, client *http.Client, baseURL string) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.token != "" && time.Until(tm.expiry) > tokenRenewSkew {
		return tm.token, nil
	}

	token, err := tm.fetch(ctx, client, baseURL)
	if err != nil {
		return "", err
	}
	tm.token = token.AccessToken
	tm.expiry = time.Now().Add(token.expiryDuration())
	return tm.token, nil
}

func (tm *tokenManager) fetch(ctx context.Context, client *http.Client, baseURL string) (*tokenResponse, error) {
	body, err := json.Marshal(map[string]string{
		"userName": tm.userName,
		"password": tm.password,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal auth body failed, %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/token", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create auth request failed, %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute auth request failed, %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read auth response failed, %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("viva auth, status %d, %s", resp.StatusCode, string(raw))
	}

	var tr tokenResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, fmt.Errorf("unmarshal token response failed, %w", err)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("viva auth, empty access_token")
	}
	return &tr, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (tr tokenResponse) expiryDuration() time.Duration {
	if tr.ExpiresIn <= 0 {
		return time.Hour
	}
	return time.Duration(tr.ExpiresIn) * time.Second
}
