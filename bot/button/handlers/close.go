package handlers

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction/component"
)

type CloseHandler struct{}

func (h *CloseHandler) Matcher() matcher.Matcher {
	return &matcher.SimpleMatcher{
		CustomId: "close",
	}
}

func (h *CloseHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

func (h *CloseHandler) Execute(ctx *context.ButtonContext) {
	// Get the ticket properties
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Check that this channel is a ticket channel
	if ticket.GuildId == 0 {
		return
	}

	closeConfirmation, err := dbclient.Client.CloseConfirmation.Get(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if closeConfirmation {
		// Make sure user can close;
		// Get user's permissions level
		permissionLevel, err := ctx.UserPermissionLevel()
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if permissionLevel == permission.Everyone {
			usersCanClose, err := dbclient.Client.UsersCanClose.Get(ctx.GuildId())
			if err != nil {
				ctx.HandleError(err)
			}

			if (permissionLevel == permission.Everyone && ticket.UserId != ctx.UserId()) || (permissionLevel == permission.Everyone && !usersCanClose) {
				ctx.Reply(constants.Red, i18n.Error, i18n.MessageCloseNoPermission)
				return
			}
		}

		// Send confirmation message
		// TODO: Translate
		confirmEmbed := utils.BuildEmbed(ctx, constants.Green, i18n.TitleCloseConfirmation, i18n.MessageCloseConfirmation, nil, ctx.PremiumTier())
		msgData := command.MessageResponse{
			Embeds: []*embed.Embed{confirmEmbed},
			Components: []component.Component{
				component.BuildActionRow(component.BuildButton(component.Button{
					Label:    ctx.GetMessage(i18n.TitleClose),
					CustomId: "close_confirm",
					Style:    component.ButtonStylePrimary,
					Emoji:    utils.BuildEmoji("✔️"),
				})),
			},
		}

		if _, err := ctx.ReplyWith(msgData); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		// TODO: IntoPanelContext()?
		logic.CloseTicket(ctx, nil, true)
	}
}
