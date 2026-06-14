package tunnel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// APIClient provides access to the independent HTTP API (port 19092).
// Authentication via appId / appKey is required for all endpoints except /api/health.
type APIClient struct {
	baseURL string
	appID   string
	appKey  string
	appName string
	client  *http.Client
}

// NewAPIClient creates a new API client.
//
// baseURL should be the full address including port, e.g. "http://192.168.1.100:19092".
// appName is the application name (e.g. "com.dustinky.qwenpaw"), used when registering domains.
func NewAPIClient(baseURL, appID, appKey, appName string, timeout time.Duration) *APIClient {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &APIClient{
		baseURL: baseURL,
		appID:   appID,
		appKey:  appKey,
		appName: appName,
		client:  &http.Client{Timeout: timeout},
	}
}

// Health performs a health check. No authentication required.
func (c *APIClient) Health() bool {
	resp, err := c.client.Get(c.baseURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var data struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false
	}
	return data.Success
}

// Status queries the tunnel running status. Authentication required.
func (c *APIClient) Status() (*TunnelStatus, error) {
	var data struct {
		Running  bool   `json:"running"`
		Status   string `json:"status"`
		PID      string `json:"pid"`
		Arch     string `json:"arch"`
		StartAt  int64  `json:"startAt"`
		TunnelID string `json:"tunnelId"`
	}
	if err := c.get("/api/status", nil, &data); err != nil {
		return nil, err
	}
	return &TunnelStatus{
		Running:  data.Running,
		Status:   data.Status,
		PID:      data.PID,
		Arch:     data.Arch,
		StartAt:  data.StartAt,
		TunnelID: data.TunnelID,
	}, nil
}

// Register registers or updates a domain forwarding rule.
//
// The appName is obtained from the client configuration set in NewAPIClient.
func (c *APIClient) Register(domain, service string) (*CGIRegisterResult, error) {
	body := struct {
		AppName string `json:"appName"`
		Domain  string `json:"domain"`
		Service string `json:"service"`
	}{AppName: c.appName, Domain: domain, Service: service}

	var data struct {
		Success  bool     `json:"success"`
		Errors   []string `json:"errors"`
		Messages []string `json:"messages"`
		Result   *struct {
			TunnelID string         `json:"tunnel_id"`
			Config   map[string]any `json:"config"`
		} `json:"result"`
	}
	if err := c.post("/api/register", body, &data); err != nil {
		return nil, err
	}
	result := &CGIRegisterResult{
		Success:  data.Success,
		Errors:   data.Errors,
		Messages: data.Messages,
	}
	if data.Result != nil {
		result.TunnelID = data.Result.TunnelID
		result.RawConfig = data.Result.Config
	}
	return result, nil
}

// DomainStatus queries the domain registration status.
//
// If appName is empty, the credentials-bound appName is used automatically.
func (c *APIClient) DomainStatus(appName string) (*DomainStatusResult, error) {
	params := url.Values{}
	if appName != "" {
		params.Set("appName", appName)
	}
	var data struct {
		Registered    bool   `json:"registered"`
		AppName       string `json:"appName"`
		Domain        string `json:"domain"`
		Service       string `json:"service"`
		DNSValid      bool   `json:"dnsValid"`
		IngressValid  bool   `json:"ingressValid"`
		TunnelRunning bool   `json:"tunnelRunning"`
		CFConfigured  bool   `json:"cfConfigured"`
		Message       string `json:"message"`
	}
	if err := c.get("/api/domain-status", params, &data); err != nil {
		return nil, err
	}
	return &DomainStatusResult{
		Registered:    data.Registered,
		AppName:       data.AppName,
		Domain:        data.Domain,
		Service:       data.Service,
		DNSValid:      data.DNSValid,
		IngressValid:  data.IngressValid,
		TunnelRunning: data.TunnelRunning,
		CFConfigured:  data.CFConfigured,
		Message:       data.Message,
	}, nil
}

// -----------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------

func (c *APIClient) authHeaders() map[string]string {
	return map[string]string{
		"X-App-Id":  c.appID,
		"X-App-Key": c.appKey,
	}
}

func (c *APIClient) get(path string, params url.Values, v any) error {
	u, _ := url.Parse(c.baseURL + path)
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	for k, v := range c.authHeaders() {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp, v)
}

func (c *APIClient) post(path string, body any, v any) error {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.authHeaders() {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp, v)
}