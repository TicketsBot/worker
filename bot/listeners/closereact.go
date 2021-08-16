package listeners

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/rest"
	"time"
)

func OnCloseReact(worker *worker.Context, data interaction.ButtonInteraction) {
	// TODO: Log this
	if data.Member == nil {
		return
	}

	// Create error context for later
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   data.GuildId.Value,
		User:    data.Member.User.Id,
		Channel: data.ChannelId,
		Shard:   worker.ShardId,
	}

	// Get user object
	user, err := worker.GetUser(data.Member.User.Id)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Ensure that the user is an actual user, not a bot
	if user.Bot {
		return
	}

	// Get the ticket properties
	ticket, err := dbclient.Client.Tickets.GetByChannel(data.ChannelId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Check that this channel is a ticket channel
	if ticket.GuildId == 0 {
		return
	}

	closeConfirmation, err := dbclient.Client.CloseConfirmation.Get(data.GuildId.Value)
	if err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	// Get whether the guild is premium
	premiumTier, err := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, worker.Token, worker.RateLimiter)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	if closeConfirmation {
		// Make sure user can close;
		// Get user's permissions level
		permissionLevel, err := permission.GetPermissionLevel(utils.ToRetriever(worker), *data.Member, data.GuildId.Value)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		if permissionLevel == permission.Everyone {
			usersCanClose, err := dbclient.Client.UsersCanClose.Get(data.GuildId.Value)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}

			if (permissionLevel == permission.Everyone && ticket.UserId != data.Member.User.Id) || (permissionLevel == permission.Everyone && !usersCanClose) {
				utils.SendEmbed(worker, data.ChannelId, data.GuildId.Value, nil, utils.Red, "Error", i18n.MessageCloseNoPermission, nil, 30, premiumTier > premium.None)
				return
			}
		}

		// Send confirmation message
		confirmEmbed := utils.BuildEmbedRaw(worker, utils.Green, "Close Confirmation", "Please confirm that you want to close the ticket", nil, premiumTier > premium.None)
		msgData := rest.CreateMessageData{
			Embeds: []*embed.Embed{confirmEmbed},
			Components: []component.Component{
				component.BuildActionRow(component.BuildButton(component.Button{
					Label:    "Close",
					CustomId: "close_confirm",
					Style:    component.ButtonStylePrimary,
					Emoji: emoji.Emoji{
						Name: "✔️",
					},
					Url:      nil,
					Disabled: false,
				})),
			},
		}

		msg, err := worker.CreateMessageComplex(data.ChannelId, msgData)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		go func() {
			time.Sleep(time.Second * 10)
			_ = worker.DeleteMessage(msg.ChannelId, msg.Id)
		}()
	} else {
		ctx := context.NewPanelContext(worker, data.GuildId.Value, data.ChannelId, data.Member.User.Id, premiumTier)
		logic.CloseTicket(&ctx, nil, true)
	}
}
