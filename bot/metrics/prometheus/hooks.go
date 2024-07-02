package prometheus

import (
	"context"
	"net/http"
	"time"
)

func PreRequestHook(_ string, req *http.Request) {
	ActiveHttpRequests.Inc()

	ctx := context.WithValue(req.Context(), "rt", time.Now())
	*req = *req.WithContext(ctx)
}

func PostRequestHook(_ string, res *http.Response) {
	ActiveHttpRequests.Dec()

	if requestTime := res.Request.Context().Value("rt"); requestTime != nil {
		duration := time.Since(requestTime.(time.Time))
		HttpRequestDuration.Observe(duration.Seconds())
	}
}
