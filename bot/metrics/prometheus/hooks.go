package prometheus

import "net/http"

func PreRequestHook(string, *http.Request) {
	ActiveHttpRequests.Inc()
}

func PostRequestHook(string, *http.Response) {
	ActiveHttpRequests.Dec()
}
