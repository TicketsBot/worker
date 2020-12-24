package event

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/go-redis/redis"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/rest/ratelimit"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type response struct {
	Success bool `json:"success"`
}

type errorResponse struct {
	response
	Error string `json:"error"`
}

func newErrorResponse(err error) errorResponse {
	return errorResponse{
		response: response{
			Success: false,
		},
		Error:    err.Error(),
	}
}

var successResponse = response{
	Success: true,
}

func HttpListen(redis *redis.Client, cache *cache.PgCache) {
	http.HandleFunc("/event", eventHandler(redis, cache))
	http.HandleFunc("/interaction", commandHandler(redis, cache))

	if err := http.ListenAndServe(os.Getenv("HTTP_ADDR"), nil); err != nil {
		panic(err)
	}
}

func eventHandler(redis *redis.Client, cache *cache.PgCache) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var event eventforwarding.Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			sentry.Error(err)

			marshalled, err := json.Marshal(newErrorResponse(err))
			if err != nil { // ???????????????????????????
				sentry.Error(err)
				return
			}

			_, _ = w.Write(marshalled)
			return
		}

		var keyPrefix string

		if event.IsWhitelabel {
			keyPrefix = fmt.Sprintf("ratelimiter:%d", event.BotId)
		} else {
			keyPrefix = "ratelimiter:public"
		}

		ctx := &worker.Context{
			Token:        event.BotToken,
			BotId:        event.BotId,
			IsWhitelabel: event.IsWhitelabel,
			ShardId:      event.ShardId,
			Cache:        cache,
			RateLimiter:  ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis, keyPrefix), 1),
		}

		marshalled, err := json.Marshal(successResponse)
		if err != nil { // ???????????????????????????
			sentry.Error(err)
			return
		}

		_, _ = w.Write(marshalled)

		if err := execute(ctx, event.Event); err != nil {
			marshalled, _ := json.Marshal(event)
			logrus.Warnf("error executing event: %v (payload: %s)", err, string(marshalled))
		}
	}
}

func commandHandler(redis *redis.Client, cache *cache.PgCache) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var command eventforwarding.Command
		if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
			sentry.Error(err)

			marshalled, err := json.Marshal(newErrorResponse(err))
			if err != nil { // ???????????????????????????
				sentry.Error(err)
				return
			}

			_, _ = w.Write(marshalled)
			return
		}

		var keyPrefix string

		if command.IsWhitelabel {
			keyPrefix = fmt.Sprintf("ratelimiter:%d", command.BotId)
		} else {
			keyPrefix = "ratelimiter:public"
		}

		ctx := &worker.Context{
			Token:        command.BotToken,
			BotId:        command.BotId,
			IsWhitelabel: command.IsWhitelabel,
			Cache:        cache,
			RateLimiter:  ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis, keyPrefix), 1),
		}

		marshalled, err := json.Marshal(successResponse)
		if err != nil { // ???????????????????????????
			sentry.Error(err)
			return
		}

		_, _ = w.Write(marshalled)

		if err := executeCommand(ctx, command.Event); err != nil {
			marshalled, _ := json.Marshal(command)
			logrus.Warnf("error executing command: %v (payload: %s)", err, string(marshalled))
		}
	}
}