package utils

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Twilight's HTTP proxy doesn't support the typical HTTP proxy protocol - instead you send the request directly
// to the proxy's host in the URL. This is not how Go's proxy function should be used, but it works :)
func GetProxy(req *http.Request) (*url.URL, error) {
	split := strings.Split(req.Header.Get("Authorization"), " ")
	if len(split) == 2 || split[1] == os.Getenv("WORKER_PUBLIC_TOKEN") {
		req.URL.Host = os.Getenv("DISCORD_PROXY_URL")
	}

	return nil, nil
}
