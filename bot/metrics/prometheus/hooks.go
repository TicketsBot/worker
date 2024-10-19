package prometheus

import (
	"context"
	"encoding/json"
	"github.com/rxdn/gdl/rest/request"
	"net/http"
	"strconv"
	"time"
)

func PreRequestHook(_ string, req *http.Request) {
	ActiveHttpRequests.Inc()

	ctx := context.WithValue(req.Context(), "rt", time.Now())
	*req = *req.WithContext(ctx)
}

func PostRequestHook(res *http.Response, body []byte) {
	ActiveHttpRequests.Dec()

	if res == nil {
		return
	}

	if requestTime := res.Request.Context().Value("rt"); requestTime != nil {
		duration := time.Since(requestTime.(time.Time))
		HttpRequestDuration.Observe(duration.Seconds())
	}

	if res.StatusCode >= 400 {
		var apiError request.ApiV8Error
		if err := json.Unmarshal(body, &apiError); err != nil {
			DiscordApiErrors.WithLabelValues(strconv.Itoa(res.StatusCode), "UNKNOWN").Inc()
			return
		}

		DiscordApiErrors.WithLabelValues(strconv.Itoa(res.StatusCode), apiError.Message).Inc()
	}
}
