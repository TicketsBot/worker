package tickets

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/permission"
)

type StartTicketCommand struct {
}

func (StartTicketCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "Start Ticket",
		Type:            interaction.ApplicationCommandTypeMessage,
		PermissionLevel: permcache.Everyone, // Customisable level
		Category:        command.Tickets,
		InteractionOnly: true,
	}
}

func (c StartTicketCommand) GetExecutor() interface{} {
	return c.Execute
}

func (StartTicketCommand) Execute(ctx registry.CommandContext) {
	interaction, ok := ctx.(*context.SlashCommandContext)
	if !ok {
		return
	}

	settings, err := dbclient.Client.Settings.Get(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	userPermissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if userPermissionLevel < permcache.PermissionLevel(settings.ContextMenuPermissionLevel) {
		ctx.Reply(utils.Red, "Error", i18n.MessageNoPermission)
		return
	}

	messageId := interaction.Interaction.Data.TargetId

	// TODO: Use resolved
	msg, err := ctx.Worker().GetChannelMessage(ctx.ChannelId(), messageId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ticket, err := logic.OpenTicket(ctx, nil, msg.Content)
	if err == nil && ticket.ChannelId != nil {
		sendTicketStartedFromMessage(ctx, ticket, msg)
		addMessageSender(ctx, ticket, msg)
	}
}

func sendTicketStartedFromMessage(ctx registry.CommandContext, ticket database.Ticket, msg message.Message)  {
	// Send info message
	isPremium := ctx.PremiumTier() >= premium.Premium

	// format
	messageLink := fmt.Sprintf("https://discord.com/channels/%d/%d/%d", ctx.GuildId(), ctx.ChannelId(), msg.Id)

	msgEmbed := utils.BuildEmbed(
		ctx.Worker(), ctx.GuildId(), utils.Green, "Ticket", i18n.MessageTicketStartedFrom, nil, isPremium,
		messageLink, msg.Author.Id, ctx.ChannelId(), utils.StringMax(msg.Content, 2048, "..."),
	)

	if _, err := ctx.Worker().CreateMessageEmbed(*ticket.ChannelId, msgEmbed); err != nil {
		ctx.HandleError(err)
		return
	}
}

func addMessageSender(ctx registry.CommandContext, ticket database.Ticket, msg message.Message)  {
	// If the sender was the ticket opener, or staff, they already have access
	// However, support teams makes this tricky
	if msg.Author.Id == ticket.UserId {
		return
	}

	// Get perms
	ch, err := ctx.Worker().GetChannel(*ticket.ChannelId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	for _, overwrite := range ch.PermissionOverwrites {
		// Check if already present
		if overwrite.Id == msg.Author.Id {
			return
		}
	}

	overwrite := channel.PermissionOverwrite{
		Id:    msg.Author.Id,
		Type:  channel.PermissionTypeMember,
		Allow: permission.BuildPermissions(logic.AllowedPermissions...),
		Deny:  0,
	}

	if err := ctx.Worker().EditChannelPermissions(*ticket.ChannelId, overwrite); err != nil {
		ctx.HandleError(err)
		return
	}
}
