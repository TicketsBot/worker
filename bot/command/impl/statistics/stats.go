package statistics

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
)

type StatsCommand struct {
}

func (StatsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "stats",
		Description:     translations.HelpStats,
		Aliases:         []string{"statistics"},
		PermissionLevel: permission.Support,
		Children: []registry.Command{
			StatsUserCommand{},
			StatsServerCommand{},
		},
		Category:    command.Statistics,
		PremiumOnly: true,
	}
}

func (c StatsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (StatsCommand) Execute(ctx registry.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!stats server`\n`t!stats @User`",
		Inline: false,
	}

	ctx.ReplyWithFields(utils.Red, "Error", translations.MessageInvalidArgument, utils.FieldsToSlice(usageEmbed))
	ctx.Reject()
}
