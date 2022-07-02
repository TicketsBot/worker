package utils

import (
	"github.com/TicketsBot/worker/config"
	"net/http"
	"strings"
)

// Twilight's HTTP proxy doesn't support the typical HTTP proxy protocol - instead you send the request directly
// to the proxy's host in the URL. This is not how Go's proxy function should be used, but it works :)
func ProxyHook(token string, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, "/api/v9/applications/") {
		req.URL.Scheme = "http"
		req.URL.Host = config.Conf.Discord.ProxyUrl
	}
}
