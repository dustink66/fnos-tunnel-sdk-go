// Package tunnel provides Go clients for the fnOS Cloudflare Tunnel API.
//
// Supports both the CGI interface (fnOS gateway proxy, no auth) and the
// independent HTTP API (port 19092, requires credentials).
//
// CGI (no auth):
//
//	client := tunnel.NewCGIClient("http://192.168.1.100", 10*time.Second)
//	status, err := client.Status()
//
// HTTP API (with auth):
//
//	client := tunnel.NewAPIClient("http://192.168.1.100:19092", "appId", "appKey", 10*time.Second)
//	healthy := client.Health()
package tunnel

import "time"

// DefaultTimeout is the default HTTP request timeout.
const DefaultTimeout = 10 * time.Second