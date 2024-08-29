package event

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button"
	btn_manager "github.com/TicketsBot/worker/bot/button/manager"
	"github.com/TicketsBot/worker/bot/command"
	cmd_manager "github.com/TicketsBot/worker/bot/command/manager"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
	"github.com/sirupsen/logrus"
	"strings"
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
	router.Use(metricsMiddleware)

	if gin.Mode() != gin.ReleaseMode {
		router.Use(gin.Logger())
	}

	// Routes
	router.POST("/event", eventHandler(cache))
	router.POST("/interaction", interactionHandler(redis, cache))

	if err := router.Run(config.Conf.Bot.HttpAddress); err != nil {
		panic(err)
	}
}

func metricsMiddleware(c *gin.Context) {
	prometheus.InboundRequests.WithLabelValues(c.Request.URL.Path).Inc()
	c.Next()
}

func eventHandler(cache *cache.PgCache) func(*gin.Context) {
	return func(c *gin.Context) {
		var event eventforwarding.Event
		if err := c.BindJSON(&event); err != nil {
			sentry.Error(err)
			c.JSON(400, newErrorResponse(err))
			return
		}

		workerCtx := &worker.Context{
			Token:        event.BotToken,
			BotId:        event.BotId,
			IsWhitelabel: event.IsWhitelabel,
			ShardId:      event.ShardId,
			Cache:        cache,
			RateLimiter:  nil, // Use http-proxy ratelimit functionality
		}

		c.AbortWithStatusJSON(200, successResponse)

		if err := execute(workerCtx, event.Event); err != nil {
			marshalled, _ := json.Marshal(event)
			logrus.Warnf("error executing event: %v (payload: %s)", err, string(marshalled))
		}
	}
}

func interactionHandler(redis *redis.Client, cache *cache.PgCache) func(*gin.Context) {
	commandManager := new(cmd_manager.CommandManager)
	commandManager.RegisterCommands()
	commandManager.RunSetupFuncs()

	buttonManager := btn_manager.NewButtonManager()
	buttonManager.RegisterCommands()

	return func(ctx *gin.Context) {
		var payload eventforwarding.Interaction
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(400, newErrorResponse(err))
			return
		}

		worker := &worker.Context{
			Token:        payload.BotToken,
			BotId:        payload.BotId,
			IsWhitelabel: payload.IsWhitelabel,
			Cache:        cache,
			RateLimiter:  nil, // Use http-proxy ratelimit functionality
		}

		switch payload.InteractionType {
		case interaction.InteractionTypeApplicationCommand:
			var interactionData interaction.ApplicationCommandInteraction
			if err := json.Unmarshal(payload.Event, &interactionData); err != nil {
				logrus.Warnf("error parsing application payload data: %v", err)
				return
			}

			responseCh := make(chan interaction.ApplicationCommandCallbackData, 1)

			deferDefault, err := executeCommand(ctx, worker, commandManager.GetCommands(), interactionData, responseCh)
			if err != nil {
				marshalled, _ := json.Marshal(payload)
				logrus.Warnf("error executing payload: %v (payload: %s)", err, string(marshalled))
				return
			}

			var flags uint
			if deferDefault {
				flags = message.SumFlags(message.FlagEphemeral)
			}

			res := interaction.NewResponseAckWithSource(flags)
			ctx.JSON(200, res)
			ctx.Writer.Flush()

			go handleApplicationCommandResponseAfterDefer(interactionData, worker, responseCh)

			prometheus.InteractionTimeToReceive.Observe(calculateTimeToReceive(interactionData.Id).Seconds())
		case interaction.InteractionTypeMessageComponent:
			var interactionData interaction.MessageComponentInteraction
			if err := json.Unmarshal(payload.Event, &interactionData); err != nil {
				logrus.Warnf("error parsing application payload data: %v", err)
				return
			}

			timeToDefer := calculateTimeToDefer(interactionData.Id)

			responseCh := make(chan button.Response, 1) // Buffer > 0 is important, or it could hang!
			btn_manager.HandleInteraction(ctx, buttonManager, worker, interactionData, responseCh)

			select {
			case <-time.After(timeToDefer):
				res := interaction.NewResponseDeferredMessageUpdate()
				ctx.JSON(200, res)
				ctx.Writer.Flush()
			case data := <-responseCh:
				ctx.JSON(200, data.Build())
				ctx.Writer.Flush()
			}

			go handleButtonResponseAfterDefer(interactionData.InteractionMetadata, worker, time.Now(), responseCh)

			prometheus.InteractionTimeToReceive.Observe(calculateTimeToReceive(interactionData.Id).Seconds())
			prometheus.InteractionTimeToDefer.Observe(timeToDefer.Seconds())
		case interaction.InteractionTypeApplicationCommandAutoComplete:
			var interactionData interaction.ApplicationCommandAutoCompleteInteraction
			if err := json.Unmarshal(payload.Event, &interactionData); err != nil {
				logrus.Warnf("error parsing application payload data: %v", err)
				return
			}

			cmd, ok := commandManager.GetCommands()[interactionData.Data.Name]
			if !ok {
				logrus.Warnf("autocomplete for invalid command: %s", interactionData.Data.Name)
				return
			}

			options := interactionData.Data.Options
			for len(options) > 0 && options[0].Value == nil { // Value and Options are mutually exclusive, value is never present on subcommands
				subCommand := options[0]

				var found bool
				for _, child := range cmd.Properties().Children {
					if child.Properties().Name == subCommand.Name {
						cmd = child
						found = true
						break
					}
				}

				if !found {
					logrus.Warnf("subcommand %s does not exist for command %s", subCommand.Name, cmd.Properties().Name)
					return
				}

				options = subCommand.Options
			}

			focused := findFocusedOption(interactionData.Data.Options)
			if focused == nil {
				logrus.Warnf("focused option not found")
				return
			}

			var handler command.AutoCompleteHandler
			for _, arg := range cmd.Properties().Arguments {
				if strings.ToLower(arg.Name) == strings.ToLower(focused.Name) {
					handler = arg.AutoCompleteHandler
				}
			}

			if handler == nil {
				logrus.Warnf("autocomplete for argument without handler: %s", focused.Name)
				return
			}

			choices := handler(interactionData, fmt.Sprintf("%v", focused.Value))
			res := interaction.NewApplicationCommandAutoCompleteResultResponse(choices)
			ctx.JSON(200, res)
			ctx.Writer.Flush()

		case interaction.InteractionTypeModalSubmit:
			var interactionData interaction.ModalSubmitInteraction
			if err := json.Unmarshal(payload.Event, &interactionData); err != nil {
				logrus.Warnf("error parsing application payload data: %v", err)
				return
			}

			ctx.JSON(200, interaction.NewResponseDeferredMessageUpdate())
			ctx.Writer.Flush()

			responseCh := make(chan button.Response, 1)
			btn_manager.HandleModalInteraction(ctx, buttonManager, worker, interactionData, responseCh)

			go handleButtonResponseAfterDefer(interactionData.InteractionMetadata, worker, time.Now(), responseCh)
		}
	}
}

func handleApplicationCommandResponseAfterDefer(interactionData interaction.ApplicationCommandInteraction, worker *worker.Context, responseCh chan interaction.ApplicationCommandCallbackData) {
	deferredAt := time.Now()
	hasReplied := false

	start := time.Now()
	prometheus.ActiveInteractions.Inc()
	defer func() {
		prometheus.ActiveInteractions.Dec()
		prometheus.InteractionTimeToComplete.Observe(time.Since(start).Seconds())
	}()

	for {
		select {
		case <-time.After(time.Second * 15):
			return
		case data, ok := <-responseCh:
			if !ok {
				return
			}

			if time.Now().Sub(utils.SnowflakeToTime(interactionData.Id)) > time.Minute*14 ||
				deferredAt.Sub(utils.SnowflakeToTime(interactionData.Id)) > config.Conf.Discord.DeferHardTimeout {
				return
			}

			if hasReplied {
				restData := rest.WebhookBody{
					Content:         data.Content,
					Embeds:          data.Embeds,
					AllowedMentions: data.AllowedMentions,
					Components:      data.Components,
				}

				if _, err := rest.CreateFollowupMessage(context.Background(), interactionData.Token, worker.RateLimiter, worker.BotId, restData); err != nil {
					sentry.ErrorWithContext(err, NewApplicationCommandInteractionErrorContext(interactionData))
					return
				}
			} else {
				hasReplied = true

				restData := rest.WebhookEditBody{
					Content:         data.Content,
					Embeds:          data.Embeds,
					AllowedMentions: data.AllowedMentions,
					Components:      data.Components,
				}

				if _, err := rest.EditOriginalInteractionResponse(context.Background(), interactionData.Token, worker.RateLimiter, worker.BotId, restData); err != nil {
					sentry.ErrorWithContext(err, NewApplicationCommandInteractionErrorContext(interactionData))
					return
				}
			}
		}
	}
}

func handleButtonResponseAfterDefer(interactionData interaction.InteractionMetadata, worker *worker.Context, deferredAt time.Time, ch chan button.Response) {
	start := time.Now()
	prometheus.ActiveInteractions.Inc()
	defer func() {
		prometheus.ActiveInteractions.Dec()
		prometheus.InteractionTimeToComplete.Observe(time.Since(start).Seconds())
	}()

	for {
		select {
		case <-time.After(time.Second * 15):
			return
		case data, ok := <-ch:
			if !ok {
				return
			}

			if time.Now().Sub(utils.SnowflakeToTime(interactionData.Id)) > time.Minute*14 ||
				deferredAt.Sub(utils.SnowflakeToTime(interactionData.Id)) > config.Conf.Discord.DeferHardTimeout {
				return
			}

			if err := data.HandleDeferred(interactionData, worker); err != nil {
				sentry.ErrorWithContext(err, NewMessageComponentInteractionErrorContext(interactionData))
			}
		}
	}
}

// TODO: Handle other data types
func findFocusedOption(options []interaction.ApplicationCommandInteractionDataOption) *interaction.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Focused {
			return &option
		}

		if option.Options != nil {
			if focused := findFocusedOption(option.Options); focused != nil {
				return focused
			}
		}
	}

	return nil
}

func calculateTimeToReceive(interactionId uint64) time.Duration {
	generated := utils.SnowflakeToTime(interactionId)
	return time.Now().Sub(generated)
}

func calculateTimeToDefer(interactionId uint64) time.Duration {
	generated := utils.SnowflakeToTime(interactionId)

	// Call max incase the snowflake timestamp is off
	return max(generated.Add(config.Conf.Discord.CallbackTimeout).Sub(time.Now()), config.Conf.Discord.CallbackTimeout)
}
