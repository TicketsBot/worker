package handlers

import (
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/customisation"
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
		Flags:   registry.SumFlags(registry.GuildAllowed),
		Timeout: constants.TimeoutCloseTicket,
	}
}

func (h *CloseHandler) Execute(ctx *cmdcontext.ButtonContext) {
	// Get the ticket properties
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx, ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Check that this channel is a ticket channel
	if ticket.GuildId == 0 {
		return
	}

	// This is checked by the close function, but we need to check before showing close confirmation
	if !utils.CanClose(ctx, ctx, ticket) {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageCloseNoPermission)
		return
	}

	closeConfirmation, err := dbclient.Client.CloseConfirmation.Get(ctx, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if closeConfirmation {
		// Send confirmation message
		confirmEmbed := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleCloseConfirmation, i18n.MessageCloseConfirmation, nil)
		confirmEmbed.SetAuthor(ctx.InteractionUser().Username, "", utils.Ptr(ctx.InteractionUser()).AvatarUrl(256))

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
		logic.CloseTicket(ctx.Context, ctx, nil, false)
	}
}
