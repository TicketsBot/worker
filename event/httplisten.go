package event

import (
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/manager"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/ratelimit"
	"github.com/sirupsen/logrus"
	"os"
	"time"
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
		Error: err.Error(),
	}
}

var successResponse = response{
	Success: true,
}

func HttpListen(redis *redis.Client, cache *cache.PgCache) {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())

	if gin.Mode() != gin.ReleaseMode {
		router.Use(gin.Logger())
	}

	// Routes
	router.POST("/event", eventHandler(redis, cache))
	router.POST("/interaction", commandHandler(redis, cache))

	if err := router.Run(os.Getenv("HTTP_ADDR")); err != nil {
		panic(err)
	}
}

func eventHandler(redis *redis.Client, cache *cache.PgCache) func(*gin.Context) {
	return func(ctx *gin.Context) {
		var event eventforwarding.Event
		if err := ctx.BindJSON(&event); err != nil {
			sentry.Error(err)
			ctx.JSON(400, newErrorResponse(err))
			return
		}

		var keyPrefix string

		if event.IsWhitelabel {
			keyPrefix = fmt.Sprintf("ratelimiter:%d", event.BotId)
		} else {
			keyPrefix = "ratelimiter:public"
		}

		workerCtx := &worker.Context{
			Token:        event.BotToken,
			BotId:        event.BotId,
			IsWhitelabel: event.IsWhitelabel,
			ShardId:      event.ShardId,
			Cache:        cache,
			RateLimiter:  ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis, keyPrefix), 1),
		}

		ctx.AbortWithStatusJSON(200, successResponse)

		if err := execute(workerCtx, event.Event); err != nil {
			marshalled, _ := json.Marshal(event)
			logrus.Warnf("error executing event: %v (payload: %s)", err, string(marshalled))
		}
	}
}

func commandHandler(redis *redis.Client, cache *cache.PgCache) func(*gin.Context) {
	commandManager := new(manager.CommandManager)
	commandManager.RegisterCommands()

	return func(ctx *gin.Context) {
		var command eventforwarding.Command
		if err := ctx.BindJSON(&command); err != nil {
			ctx.JSON(400, newErrorResponse(err))
			return
		}

		var keyPrefix string

		if command.IsWhitelabel {
			keyPrefix = fmt.Sprintf("ratelimiter:%d", command.BotId)
		} else {
			keyPrefix = "ratelimiter:public"
		}

		workerCtx := &worker.Context{
			Token:        command.BotToken,
			BotId:        command.BotId,
			IsWhitelabel: command.IsWhitelabel,
			Cache:        cache,
			RateLimiter:  ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis, keyPrefix), 1),
		}

		var interactionData interaction.ApplicationCommandInteraction
		if err := json.Unmarshal(command.Event, &interactionData); err != nil {
			logrus.Warnf("error parsing application command data: %v", err)
			return
		}

		responseCh := make(chan interaction.ApplicationCommandCallbackData, 1)

		deferDefault, err := executeCommand(workerCtx, commandManager.GetCommands(), interactionData, responseCh)
		if err != nil {
			marshalled, _ := json.Marshal(command)
			logrus.Warnf("error executing command: %v (payload: %s)", err, string(marshalled))
			return
		}

		timeout := time.NewTimer(time.Second * 2)

		select {
		case <-timeout.C:
			var flags uint
			if deferDefault {
				flags = message.SumFlags(message.FlagEphemeral)
			}

			res := interaction.NewResponseAckWithSource(flags)
			ctx.JSON(200, res)
			ctx.Writer.Flush()

			go handleResponseAfterDefer(interactionData, workerCtx, responseCh)
		case data := <-responseCh:
			res := interaction.NewResponseChannelMessage(data)
			ctx.JSON(200, res)
		}
	}
}

func handleResponseAfterDefer(interactionData interaction.ApplicationCommandInteraction, workerCtx *worker.Context, responseCh chan interaction.ApplicationCommandCallbackData) {
	timeout := time.NewTimer(time.Second * 15)

	select {
	case <-timeout.C:
		return
	case data := <-responseCh:
		restData := rest.WebhookBody{
			Content:         data.Content,
			Tts:             data.Tts,
			Flags:           data.Flags,
			Embeds:          data.Embeds,
			AllowedMentions: data.AllowedMentions,
		}

		if _, err := rest.ExecuteWebhook(interactionData.Token, workerCtx.RateLimiter, workerCtx.BotId, false, restData); err != nil {
			sentry.LogWithContext(err, buildErrorContext(interactionData))
			return
		}
	}
}

func buildErrorContext(data interaction.ApplicationCommandInteraction) sentry.ErrorContext {
	var userId uint64
	if data.User != nil {
		userId = data.User.Id
	} else if data.Member != nil {
		userId = data.Member.User.Id
	}

	return errorcontext.WorkerErrorContext{
		Guild:   data.GuildId.Value,
		User:    userId,
		Channel: data.ChannelId,
	}
}
