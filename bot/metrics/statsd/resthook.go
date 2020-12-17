package statsd

import "net/http"

func RestHook(string, *http.Request) {
	go Client.IncrementKey(KeyRest)
}
