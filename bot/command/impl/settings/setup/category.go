package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"strings"
)

type CategorySetupCommand struct{}

func (CategorySetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "category",
		Description:     translations.HelpSetup,
		Aliases:         []string{"ticketcategory", "cat", "channelcategory"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (CategorySetupCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply(utils.Red, "Setup", translations.SetupCategoryInvalid)
		ctx.ReactWithCross()
		return
	}

	name := strings.Join(ctx.Args, " ")
	channels, err := ctx.Worker.GetGuildChannels(ctx.GuildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	var category channel.Channel
	for _, ch := range channels {
		if ch.Type == channel.ChannelTypeGuildCategory && strings.EqualFold(ch.Name, name) {
			category = ch
			break
		}
	}

	if category.Id == 0 {
		ctx.Reply(utils.Red, "Setup", translations.SetupCategoryInvalid)
		ctx.ReactWithCross()
		return
	}

	if err := dbclient.Client.ChannelCategory.Set(ctx.GuildId, category.Id); err == nil {
		ctx.ReactWithCheck()
		ctx.Reply(utils.Green, "Setup", translations.SetupCategoryComplete, category.Name)
	} else {
		ctx.HandleError(err)
	}
}
