package statistics

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
)

type StatsCommand struct {
}

func (StatsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "stats",
		Description:     i18n.HelpStats,
		Type:            interaction.ApplicationCommandTypeChatInput,
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

	ctx.ReplyWithFields(constants.Red, "Error", i18n.MessageInvalidArgument, utils.FieldsToSlice(usageEmbed))
	ctx.Reject()
}
