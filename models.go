package tunnel

// TunnelStatus represents the running status of a Cloudflare Tunnel.
type TunnelStatus struct {
	// Running indicates whether the tunnel process is running.
	Running bool `json:"running"`
	// Status is the health status: "healthy" or "down".
	Status string `json:"status"`
	// PID is the process ID of the tunnel daemon.
	PID string `json:"pid,omitempty"`
	// Arch is the device architecture: "amd64" or "arm64".
	Arch string `json:"arch,omitempty"`
	// StartAt is the Unix timestamp (seconds) when the tunnel started.
	StartAt int64 `json:"startAt,omitempty"`
	// TunnelID is the ID of the currently running tunnel.
	TunnelID string `json:"tunnelId,omitempty"`
}

// DomainRegistration holds the parameters for registering a domain.
type DomainRegistration struct {
	// AppName is the unique application identifier (e.g. "com.dustinky.qwenpaw").
	AppName string `json:"appName"`
	// Domain is the full domain name (e.g. "qwenpaw.example.com").
	Domain string `json:"domain"`
	// Service is the local service URL (e.g. "http://localhost:19091").
	Service string `json:"service"`
}

// CGIRegisterResult is the result of a domain registration via CGI.
type CGIRegisterResult struct {
	// Success indicates whether the registration was successful.
	Success bool `json:"success"`
	// TunnelID is the ID of the tunnel.
	TunnelID string `json:"-"`
	// Errors contains any error messages from Cloudflare.
	Errors []string `json:"errors"`
	// Messages contains informational messages.
	Messages []string `json:"messages"`
	// RawConfig is the raw ingress config returned by Cloudflare.
	RawConfig map[string]any `json:"-"`
}

// DomainStatusResult is the result of querying a domain's registration status.
type DomainStatusResult struct {
	// Registered indicates whether the domain has been registered.
	Registered bool `json:"registered"`
	// AppName is the application name.
	AppName string `json:"appName"`
	// Domain is the registered domain name.
	Domain string `json:"domain,omitempty"`
	// Service is the registered local service address.
	Service string `json:"service,omitempty"`
	// DNSValid indicates whether the DNS CNAME record exists.
	DNSValid bool `json:"dnsValid,omitempty"`
	// IngressValid indicates whether the ingress rule is still in the config.
	IngressValid bool `json:"ingressValid,omitempty"`
	// TunnelRunning indicates whether the tunnel process is running.
	TunnelRunning bool `json:"tunnelRunning"`
	// CFConfigured indicates whether Cloudflare credentials are configured.
	CFConfigured bool `json:"cfConfigured"`
	// Message is an optional informational message.
	Message string `json:"message,omitempty"`
}

// APIError represents an error returned by the API.
type APIError struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "unknown API error"
}