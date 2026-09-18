package utils

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Placeholder base URL used for requests sent over a unix domain socket,
// the host is ignored by the dialer but still required to build a valid URL.
const SocketBaseURL = "http://localhost"

var socketTransports sync.Map

// UnixSocketPath returns the socket path when addr refers to a unix domain
// socket, either as a unix:// URL or as an absolute filesystem path.
func UnixSocketPath(addr string) (string, bool) {
	if strings.HasPrefix(addr, "unix://") {
		return strings.TrimPrefix(addr, "unix://"), true
	}
	if strings.HasPrefix(addr, "unix:") {
		return strings.TrimPrefix(addr, "unix:"), true
	}
	if strings.HasPrefix(addr, "/") {
		return addr, true
	}
	return "", false
}

// UnixDialer returns a dial function that always connects to the given socket.
func UnixDialer(path string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	var dialer net.Dialer
	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, "unix", path)
	}
}

// UnixTransport returns a shared HTTP transport that connects to the given socket.
func UnixTransport(path string) http.RoundTripper {
	if t, ok := socketTransports.Load(path); ok {
		return t.(http.RoundTripper)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = UnixDialer(path)

	t, _ := socketTransports.LoadOrStore(path, transport)
	return t.(http.RoundTripper)
}
