package setup

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type CategorySetupCommand struct{}

func (CategorySetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "category",
		Description:     i18n.HelpSetup,
		Aliases:         []string{"ticketcategory", "cat", "channelcategory"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgumentInteractionOnly("category", "Channel category for tickets to be created under", interaction.OptionTypeChannel, i18n.SetupCategoryInvalid),
			command.NewRequiredArgumentMessageOnly("category", "Name of the channel category", interaction.OptionTypeString, i18n.SetupCategoryInvalid),
		),
	}
}

func (c CategorySetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (CategorySetupCommand) Execute(ctx registry.CommandContext, categoryId *uint64, categoryName *string) {
	var category channel.Channel
	if categoryId != nil {
		var err error
		category, err = ctx.Worker().GetChannel(*categoryId)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if category.Type != channel.ChannelTypeGuildCategory {
			ctx.Reply(utils.Red, "Error", i18n.SetupCategoryInvalid)
			ctx.Reject()
			return
		}
	} else if categoryName != nil {
		channels, err := ctx.Worker().GetGuildChannels(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		var found bool
		for _, ch := range channels {
			if ch.Type == channel.ChannelTypeGuildCategory && strings.EqualFold(ch.Name, *categoryName) {
				category = ch
				found = true
				break
			}
		}

		if !found {
			ctx.Reply(utils.Red, "Setup", i18n.SetupCategoryInvalid)
			ctx.Reject()
			return
		}
	} else { // Should not be possible
		ctx.Reply(utils.Red, "Setup", i18n.SetupCategoryInvalid)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.ChannelCategory.Set(ctx.GuildId(), category.Id); err == nil {
		ctx.Accept()
		ctx.Reply(utils.Green, "Setup", i18n.SetupCategoryComplete, category.Name)
	} else {
		ctx.HandleError(err)
	}
}
