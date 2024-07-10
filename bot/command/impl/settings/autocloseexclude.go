package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type AutoCloseExcludeCommand struct {
}

func (AutoCloseExcludeCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "exclude",
		Description:      i18n.HelpAutoCloseExclude,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Support,
		Category:         command.Settings,
		DefaultEphemeral: true,
		Timeout:          time.Second * 5,
	}
}

func (c AutoCloseExcludeCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AutoCloseExcludeCommand) Execute(ctx registry.CommandContext) {
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx, ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.Id == 0 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	if err := dbclient.Client.AutoCloseExclude.Exclude(ctx, ctx.GuildId(), ticket.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(customisation.Green, i18n.TitleAutoclose, i18n.MessageAutoCloseExclude)
}
