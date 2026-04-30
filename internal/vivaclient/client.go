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

type Client struct {
	http   *http.Client
	cfg    Config
	mu     sync.Mutex
	token  string
	expiry time.Time
}

func New(cfg Config) *Client {
	return &Client{http: http.DefaultClient, cfg: cfg}
}

func (c *Client) base() string {
	return strings.TrimRight(strings.TrimSpace(c.cfg.BaseURL), "/")
}

func (c *Client) bearer(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Until(c.expiry) > renewSkew {
		return c.token, nil
	}
	body, err := json.Marshal(map[string]string{
		"userName": c.cfg.UserName,
		"password": c.cfg.Password,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/auth/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("viva auth: status %d: %s", resp.StatusCode, string(raw))
	}
	var tr tokenResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", err
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("viva auth: empty access_token")
	}
	c.token = tr.AccessToken
	ttl := time.Duration(tr.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = time.Hour
	}
	c.expiry = time.Now().Add(ttl)
	return c.token, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	tok, err := c.bearer(ctx)
	if err != nil {
		return err
	}
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base()+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("viva %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// Viva биндит phoneNum / productName (и otp) из query string, не из JSON body — см. docs/bor.json.
func subscriptionPostPath(path, phoneNum, productName string, otp *string) string {
	q := url.Values{}
	q.Set("phoneNum", strings.TrimSpace(phoneNum))
	q.Set("productName", strings.TrimSpace(productName))
	if otp != nil {
		q.Set("otp", strings.TrimSpace(*otp))
	}
	return path + "?" + q.Encode()
}

func (c *Client) doPostQuery(ctx context.Context, pathWithQuery string, out any) error {
	tok, err := c.bearer(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+pathWithQuery, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("viva POST %s: status %d: %s", strings.Split(pathWithQuery, "?")[0], resp.StatusCode, string(raw))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) GetSubscriberInfo(ctx context.Context, msisdn string) (*GetSubInfoResponse, error) {
	tok, err := c.bearer(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/Subscriber/%s", strings.TrimSpace(msisdn))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("viva subscriber not found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("viva GET subscriber: status %d: %s", resp.StatusCode, string(raw))
	}
	var out GetSubInfoResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) InitSubscription(ctx context.Context, phoneNum, productName string) (*ResponseModel, error) {
	path := subscriptionPostPath("/api/Subscription/InitSubscription", phoneNum, productName, nil)
	var out ResponseModel
	if err := c.doPostQuery(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ConfirmSubscription(ctx context.Context, phoneNum, productName string, otp *string) (*ResponseModel, error) {
	path := subscriptionPostPath("/api/Subscription/ConfirmSubscription", phoneNum, productName, otp)
	var out ResponseModel
	if err := c.doPostQuery(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RemoveSubscription(ctx context.Context, phoneNum, productName string) (*ResponseModel, error) {
	path := subscriptionPostPath("/api/Subscription/RemoveSubscription", phoneNum, productName, nil)
	var out ResponseModel
	if err := c.doPostQuery(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetProductsByPhoneNum(ctx context.Context, phoneNum string) (*ExtAppProductsRespDTO, error) {
	tok, err := c.bearer(ctx)
	if err != nil {
		return nil, err
	}
	path := "/api/Subscription/GetProductsByPhoneNum?phoneNum=" + strings.TrimSpace(phoneNum)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("viva GetProductsByPhoneNum: status %d: %s", resp.StatusCode, string(raw))
	}
	var env extAppProductsEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if env.Result == nil {
		return &ExtAppProductsRespDTO{PhoneNumber: phoneNum}, nil
	}
	return env.Result, nil
}
