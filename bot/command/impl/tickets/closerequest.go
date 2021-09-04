package tickets

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/rest"
	"strings"
	"time"
)

type CloseRequestCommand struct {
}

func (CloseRequestCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "closerequest",
		Description:     i18n.HelpCloseRequest,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
		InteractionOnly: true,
		Arguments: command.Arguments(
			command.NewOptionalArgument("close_delay", "Hours to close the ticket in if the user does not respond", interaction.OptionTypeInteger, "infallible"),
			command.NewOptionalArgument("reason", "The reason the ticket was closed", interaction.OptionTypeString, "infallible"),
		),
	}
}

func (c CloseRequestCommand) GetExecutor() interface{} {
	return c.Execute
}

func (CloseRequestCommand) Execute(ctx registry.CommandContext, closeDelay *int, reason *string) {
	var interaction interaction.ApplicationCommandInteraction
	{
		v, ok := ctx.(*context.SlashCommandContext)
		if !ok {
			return
		}

		interaction = v.Interaction
	}

	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.Id == 0 {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	if reason != nil && len(*reason) > 255 {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageCloseReasonTooLong)
		return
	}

	var closeAt *time.Time = nil
	if closeDelay != nil {
		tmp := time.Now().Add(time.Hour * time.Duration(*closeDelay))
		closeAt = &tmp
	}

	closeRequest := database.CloseRequest{
		GuildId:  ticket.GuildId,
		TicketId: ticket.Id,
		UserId:   ctx.UserId(),
		CloseAt:  closeAt,
		Reason:   reason,
	}

	if err := dbclient.Client.CloseRequest.Set(closeRequest); err != nil {
		ctx.HandleError(err)
		return
	}

	var messageId i18n.MessageId
	var format []interface{}
	if reason == nil {
		messageId = i18n.MessageCloseRequestNoReason
		format = []interface{}{ctx.UserId()}
	} else {
		messageId = i18n.MessageCloseRequestWithReason
		format = []interface{}{ctx.UserId(), strings.ReplaceAll(*reason, "`", "\\`")}
	}

	msgEmbed := utils.BuildEmbed(ctx, constants.Green, "Close Request", messageId, nil, format...)
	components := component.BuildActionRow(
		component.BuildButton(component.Button{
			Label:    "Accept Close Request",
			CustomId: "close_request_accept",
			Style:    component.ButtonStyleSuccess,
			Emoji:    utils.BuildEmoji("☑️"),
		}),

		component.BuildButton(component.Button{
			Label:    "Deny Close Request",
			CustomId: "close_request_deny",
			Style:    component.ButtonStyleSecondary,
			Emoji:    utils.BuildEmoji("❌"),
		}),
	)

	data := command.MessageResponse{
		Content: fmt.Sprintf("<@%d>", ticket.UserId),
		Embeds:  []*embed.Embed{msgEmbed},
		AllowedMentions: message.AllowedMention{
			Users: []uint64{ticket.UserId},
		},
		Components: []component.Component{components},
	}

	msg, err := rest.ExecuteWebhook(interaction.Token, ctx.Worker().RateLimiter, interaction.ApplicationId, true, data.IntoWebhookBody())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if err := dbclient.Client.CloseRequest.SetMessageId(ticket.GuildId, ticket.Id, msg.Id); err != nil {
		ctx.HandleError(err)
		return
	}
}
