package setup

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/interaction"
)

type CategorySetupCommand struct{}

func (CategorySetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "category",
		Description:     i18n.HelpSetup,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"ticketcategory", "cat", "channelcategory"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		InteractionOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgumentInteractionOnly("category", "Name of the channel category", interaction.OptionTypeChannel, i18n.SetupCategoryInvalid),
		),
	}
}

func (c CategorySetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (CategorySetupCommand) Execute(ctx registry.CommandContext, channelId uint64) {
	category, err := ctx.Worker().GetChannel(channelId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if category.Type != channel.ChannelTypeGuildCategory {
		ctx.Reply(constants.Red, "Error", i18n.SetupCategoryInvalid)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.ChannelCategory.Set(ctx.GuildId(), category.Id); err == nil {
		ctx.Accept()
		ctx.Reply(constants.Green, "Setup", i18n.SetupCategoryComplete, category.Name)
	} else {
		ctx.HandleError(err)
	}
}
