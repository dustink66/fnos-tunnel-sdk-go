package tunnel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const cgiPath = "/cgi/ThirdParty/com.dustinky.tunnel/api.cgi"

// CGIClient provides access to the CGI interface via the fnOS gateway proxy.
// No authentication is required.
type CGIClient struct {
	baseURL string
	client  *http.Client
}

// NewCGIClient creates a new CGI client.
//
// baseURL should be the fnOS device address, e.g. "http://192.168.1.100".
func NewCGIClient(baseURL string, timeout time.Duration) *CGIClient {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &CGIClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

// Status queries the tunnel running status.
func (c *CGIClient) Status() (*TunnelStatus, error) {
	var data struct {
		Running   bool   `json:"running"`
		Status    string `json:"status"`
		PID       string `json:"pid"`
		Arch      string `json:"arch"`
		StartAt   int64  `json:"startAt"`
		TunnelID  string `json:"tunnelId"`
	}
	if err := c.get("status", nil, &data); err != nil {
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

// RegisterAppDomain registers or updates a domain forwarding rule.
//
// appName is the unique application identifier.
// domain is the full domain name.
// service is the local service URL.
func (c *CGIClient) RegisterAppDomain(appName, domain, service string) (*CGIRegisterResult, error) {
	body := DomainRegistration{
		AppName: appName,
		Domain:  domain,
		Service: service,
	}
	var data struct {
		Success  bool     `json:"success"`
		Errors   []string `json:"errors"`
		Messages []string `json:"messages"`
		Result   *struct {
			TunnelID string         `json:"tunnel_id"`
			Config   map[string]any `json:"config"`
		} `json:"result"`
	}
	if err := c.post("register_app_domain", body, &data); err != nil {
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

// GetAppDomainStatus queries the domain registration status for an application.
func (c *CGIClient) GetAppDomainStatus(appName string) (*DomainStatusResult, error) {
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
	if err := c.get("get_app_domain_status", url.Values{"appName": {appName}}, &data); err != nil {
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

func (c *CGIClient) get(action string, params url.Values, v any) error {
	u, _ := url.Parse(c.baseURL + cgiPath)
	q := u.Query()
	q.Set("action", action)
	for k, vals := range params {
		for _, val := range vals {
			q.Add(k, val)
		}
	}
	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp, v)
}

func (c *CGIClient) post(action string, body any, v any) error {
	u, _ := url.Parse(c.baseURL + cgiPath)
	q := u.Query()
	q.Set("action", action)
	u.RawQuery = q.Encode()

	b, _ := json.Marshal(body)
	resp, err := c.client.Post(u.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp, v)
}

func handleResponse(resp *http.Response, v any) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Check for API-level error first
	var errCheck struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &errCheck); err != nil {
		return fmt.Errorf("invalid JSON response: %s", string(raw[:min(len(raw), 200)]))
	}
	if !errCheck.Success && errCheck.Message != "" {
		return &APIError{Success: false, Message: errCheck.Message}
	}

	if v != nil {
		if err := json.Unmarshal(raw, v); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}