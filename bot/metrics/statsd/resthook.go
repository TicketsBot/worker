package statsd

import "net/http"

func RestHook(string, *http.Request) {
	Client.IncrementKey(KeyRest)
}
