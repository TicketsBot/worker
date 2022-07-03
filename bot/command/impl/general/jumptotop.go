package general

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
)

type JumpToTopCommand struct {
}

func (JumpToTopCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "jumptotop",
		Description:      i18n.HelpJumpToTop,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Everyone,
		Category:         command.General,
		DefaultEphemeral: true,
	}
}

func (c JumpToTopCommand) GetExecutor() interface{} {
	return c.Execute
}

func (JumpToTopCommand) Execute(ctx registry.CommandContext) {
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.Id == 0 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	if ticket.WelcomeMessageId == nil {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageJumpToTopNoWelcomeMessage)
		return
	}

	messageLink := fmt.Sprintf("https://discord.com/channels/%d/%d/%d", ctx.GuildId(), ctx.ChannelId(), *ticket.WelcomeMessageId)

	res := command.NewEphemeralEmbedMessageResponse(utils.BuildEmbed(ctx, customisation.Green, i18n.TitleAbout, i18n.MessageAbout, nil))
	res.Components = []component.Component{
		component.BuildActionRow(component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.ClickHere),
			Style:    component.ButtonStyleLink,
			Emoji:    nil,
			Url:      utils.Ptr(messageLink),
			Disabled: false,
		})),
	}

	if _, err := ctx.ReplyWith(res); err != nil {
		ctx.HandleError(err)
		return
	}
}
