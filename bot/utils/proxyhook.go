package utils

import (
	"net/http"
	"os"
	"strings"
)

// Twilight's HTTP proxy doesn't support the typical HTTP proxy protocol - instead you send the request directly
// to the proxy's host in the URL. This is not how Go's proxy function should be used, but it works :)
func ProxyHook(token string, req *http.Request) {
	if token == os.Getenv("WORKER_PUBLIC_TOKEN") {
		if !strings.HasPrefix(req.URL.Path, "/api/v8/applications/") {
			req.URL.Scheme = "http"
			req.URL.Host = os.Getenv("DISCORD_PROXY_URL")
		}
	}
}
