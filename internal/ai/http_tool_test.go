package ai

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHttpRequestHandler(t *testing.T) {
	t.Run("creates handler with configured client", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.client)
		assert.Equal(t, httpRequestTimeout, handler.client.Timeout)
	})
}

func TestHttpRequestHandler_Execute(t *testing.T) {
	t.Run("rejects non-GET methods", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		methods := []string{"POST", "PUT", "DELETE", "PATCH"}
		for _, method := range methods {
			result := handler.Execute(ctx, "https://example.com", method, []string{"example.com"})
			assert.False(t, result.Success)
			assert.Contains(t, result.Error, "Only GET method is supported")
		}
	})

	t.Run("rejects invalid URLs", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, "://invalid-url", "GET", []string{"*"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Invalid URL")
	})

	t.Run("rejects non-http/https schemes", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		schemes := []string{"ftp://example.com", "file:///etc/passwd", "javascript:alert(1)"}
		for _, url := range schemes {
			result := handler.Execute(ctx, url, "GET", []string{"*"})
			assert.False(t, result.Success)
			assert.Contains(t, result.Error, "http or https scheme")
		}
	})

	t.Run("rejects URLs without hostname", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, "https:///path", "GET", []string{"*"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "must have a hostname")
	})

	t.Run("rejects HTTP for non-localhost", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, "http://example.com/api", "GET", []string{"example.com"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "HTTPS is required")
	})

	t.Run("allows HTTP for localhost", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, server.URL, "GET", []string{"localhost", "127.0.0.1"})
		assert.True(t, result.Success)
	})

	t.Run("rejects URLs with embedded credentials", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, "https://user:pass@example.com", "GET", []string{"example.com"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "embedded credentials are not allowed")
	})

	t.Run("rejects when no domains are allowed", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, "https://example.com", "GET", []string{})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "No HTTP domains are allowed")
	})

	t.Run("rejects domains not in whitelist", func(t *testing.T) {
		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, "https://evil.com", "GET", []string{"example.com", "api.example.com"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "not in the allowed domains list")
		assert.Equal(t, []string{"example.com", "api.example.com"}, result.AllowedDomains)
	})

	t.Run("handles successful JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, httpUserAgent, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Hello",
				"count":   42,
			})
		}))
		defer server.Close()

		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, server.URL, "GET", []string{"localhost", "127.0.0.1"})
		assert.True(t, result.Success)
		assert.Equal(t, 200, result.Status)

		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Hello", data["message"])
		assert.Equal(t, float64(42), data["count"])
	})

	t.Run("rejects non-JSON content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html>Not JSON</html>"))
		}))
		defer server.Close()

		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, server.URL, "GET", []string{"localhost", "127.0.0.1"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Only JSON responses are supported")
	})

	t.Run("handles non-2xx status codes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		}))
		defer server.Close()

		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, server.URL, "GET", []string{"localhost", "127.0.0.1"})
		assert.False(t, result.Success)
		assert.Equal(t, 404, result.Status)
		assert.Contains(t, result.Error, "HTTP 404")
	})

	t.Run("handles invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		handler := NewHttpRequestHandler()
		ctx := context.Background()

		result := handler.Execute(ctx, server.URL, "GET", []string{"localhost", "127.0.0.1"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Failed to parse JSON")
	})

	t.Run("handles request timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		handler := NewHttpRequestHandler()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		result := handler.Execute(ctx, server.URL, "GET", []string{"localhost", "127.0.0.1"})
		assert.False(t, result.Success)
		// Either timeout or context deadline exceeded
		assert.True(t, result.Error != "" || !result.Success)
	})
}

func TestIsDomainAllowed(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		assert.True(t, isDomainAllowed("example.com", []string{"example.com"}))
		assert.True(t, isDomainAllowed("api.example.com", []string{"api.example.com"}))
	})

	t.Run("case insensitive match", func(t *testing.T) {
		assert.True(t, isDomainAllowed("EXAMPLE.COM", []string{"example.com"}))
		assert.True(t, isDomainAllowed("example.com", []string{"EXAMPLE.COM"}))
	})

	t.Run("wildcard subdomain match", func(t *testing.T) {
		allowedDomains := []string{"*.example.com"}
		assert.True(t, isDomainAllowed("api.example.com", allowedDomains))
		assert.True(t, isDomainAllowed("sub.api.example.com", allowedDomains))
		assert.True(t, isDomainAllowed("example.com", allowedDomains))
	})

	t.Run("no match for different domain", func(t *testing.T) {
		assert.False(t, isDomainAllowed("evil.com", []string{"example.com"}))
		assert.False(t, isDomainAllowed("example.org", []string{"example.com"}))
	})

	t.Run("no match for subdomain without wildcard", func(t *testing.T) {
		assert.False(t, isDomainAllowed("sub.example.com", []string{"example.com"}))
	})

	t.Run("handles empty allowed domains", func(t *testing.T) {
		assert.False(t, isDomainAllowed("example.com", []string{}))
	})

	t.Run("handles empty strings in allowed domains", func(t *testing.T) {
		assert.False(t, isDomainAllowed("example.com", []string{"", "  "}))
		assert.True(t, isDomainAllowed("example.com", []string{"", "example.com"}))
	})
}

func TestValidateNotPrivateIP(t *testing.T) {
	t.Run("blocks localhost", func(t *testing.T) {
		err := validateNotPrivateIP("localhost")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "localhost is not allowed")
	})

	t.Run("blocks ip6-localhost", func(t *testing.T) {
		err := validateNotPrivateIP("ip6-localhost")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "localhost is not allowed")
	})

	t.Run("blocks metadata.google.internal", func(t *testing.T) {
		err := validateNotPrivateIP("metadata.google.internal")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal hostname")
	})

	t.Run("blocks .local domains", func(t *testing.T) {
		err := validateNotPrivateIP("server.local")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "local domain")
	})

	t.Run("blocks kubernetes internal domains", func(t *testing.T) {
		err := validateNotPrivateIP("kubernetes.default.svc.cluster.local")
		require.Error(t, err)
	})

	t.Run("blocks metadata subdomains", func(t *testing.T) {
		err := validateNotPrivateIP("sub.metadata.google.internal")
		require.Error(t, err)
	})
}

func TestIsPrivateIPAddress(t *testing.T) {
	t.Run("returns false for nil IP", func(t *testing.T) {
		assert.False(t, isPrivateIPAddress(nil))
	})

	t.Run("returns true for loopback addresses", func(t *testing.T) {
		assert.True(t, isPrivateIPAddress(net.ParseIP("127.0.0.1")))
		assert.True(t, isPrivateIPAddress(net.ParseIP("127.0.0.255")))
		assert.True(t, isPrivateIPAddress(net.ParseIP("::1")))
	})

	t.Run("returns true for private RFC1918 addresses", func(t *testing.T) {
		privateIPs := []string{
			"10.0.0.1",
			"10.255.255.255",
			"172.16.0.1",
			"172.31.255.255",
			"192.168.0.1",
			"192.168.255.255",
		}
		for _, ip := range privateIPs {
			assert.True(t, isPrivateIPAddress(net.ParseIP(ip)), "Expected %s to be private", ip)
		}
	})

	t.Run("returns true for link-local addresses", func(t *testing.T) {
		assert.True(t, isPrivateIPAddress(net.ParseIP("169.254.0.1")))
		assert.True(t, isPrivateIPAddress(net.ParseIP("169.254.169.254"))) // AWS metadata
	})

	t.Run("returns true for carrier-grade NAT addresses", func(t *testing.T) {
		assert.True(t, isPrivateIPAddress(net.ParseIP("100.64.0.1")))
		assert.True(t, isPrivateIPAddress(net.ParseIP("100.127.255.255")))
	})

	t.Run("returns true for multicast addresses", func(t *testing.T) {
		assert.True(t, isPrivateIPAddress(net.ParseIP("224.0.0.1")))
		assert.True(t, isPrivateIPAddress(net.ParseIP("239.255.255.255")))
	})

	t.Run("returns false for public IP addresses", func(t *testing.T) {
		publicIPs := []string{
			"8.8.8.8",
			"1.1.1.1",
			"93.184.216.34",  // example.com
			"142.250.190.46", // google.com
		}
		for _, ip := range publicIPs {
			assert.False(t, isPrivateIPAddress(net.ParseIP(ip)), "Expected %s to be public", ip)
		}
	})

	t.Run("returns true for IPv6 unique local addresses", func(t *testing.T) {
		assert.True(t, isPrivateIPAddress(net.ParseIP("fc00::1")))
		assert.True(t, isPrivateIPAddress(net.ParseIP("fd00::1")))
	})

	t.Run("returns true for IPv6 link-local addresses", func(t *testing.T) {
		assert.True(t, isPrivateIPAddress(net.ParseIP("fe80::1")))
	})

	t.Run("returns true for TEST-NET addresses", func(t *testing.T) {
		assert.True(t, isPrivateIPAddress(net.ParseIP("192.0.2.1")))    // TEST-NET-1
		assert.True(t, isPrivateIPAddress(net.ParseIP("198.51.100.1"))) // TEST-NET-2
		assert.True(t, isPrivateIPAddress(net.ParseIP("203.0.113.1")))  // TEST-NET-3
	})
}

func TestHttpRequestResult_Struct(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := HttpRequestResult{
			Success: true,
			Status:  200,
			Data:    map[string]string{"key": "value"},
		}

		assert.True(t, result.Success)
		assert.Equal(t, 200, result.Status)
		assert.NotNil(t, result.Data)
	})

	t.Run("error result", func(t *testing.T) {
		result := HttpRequestResult{
			Success:        false,
			Error:          "Domain not allowed",
			AllowedDomains: []string{"example.com"},
		}

		assert.False(t, result.Success)
		assert.Equal(t, "Domain not allowed", result.Error)
		assert.Equal(t, []string{"example.com"}, result.AllowedDomains)
	})
}
