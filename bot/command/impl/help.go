package impl

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/elliotchance/orderedmap"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type HelpCommand struct {
}

func (HelpCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "help",
		Description:     translations.HelpHelp,
		Aliases:         []string{"h"},
		PermissionLevel: permission.Everyone,
		Category:        command.General,
	}
}

func (c HelpCommand) GetExecutor() interface{} {
	return c.Execute
}

func (h HelpCommand) Execute(ctx command.CommandContext) {
	commandCategories := orderedmap.NewOrderedMap()

	// initialise map with the correct order of categories
	for _, category := range command.Categories {
		commandCategories.Set(category, nil)
	}

	for _, cmd := range Commands {
		// check bot admin / helper only commands
		if (cmd.Properties().AdminOnly && !utils.IsBotAdmin(ctx.Author.Id)) || (cmd.Properties().HelperOnly && !utils.IsBotHelper(ctx.Author.Id)) {
			continue
		}

		if ctx.UserPermissionLevel >= cmd.Properties().PermissionLevel { // only send commands the user has permissions for
			var current []command.Command
			if commands, ok := commandCategories.Get(cmd.Properties().Category); ok {
				if commands == nil {
					current = make([]command.Command, 0)
				} else {
					current = commands.([]command.Command)
				}
			}
			current = append(current, cmd)

			commandCategories.Set(cmd.Properties().Category, current)
		}
	}

	// get prefix
	prefix := getPrefix(ctx.GuildId)

	embed := embed.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Help")

	for _, category := range commandCategories.Keys() {
		var commands []command.Command
		if retrieved, ok := commandCategories.Get(category.(command.Category)); ok {
			if retrieved == nil {
				commands = make([]command.Command, 0)
			} else {
				commands = retrieved.([]command.Command)
			}
		}

		if len(commands) > 0 {
			formatted := make([]string, 0)
			for _, cmd := range commands {
				formatted = append(formatted, command.FormatHelp(cmd, ctx.GuildId, prefix))
			}

			embed.AddField(string(category.(command.Category)), strings.Join(formatted, "\n"), false)
		}
	}

	dmChannel, err := ctx.Worker.CreateDM(ctx.Author.Id)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	if ctx.PremiumTier == premium.None {
		self, _ := ctx.Worker.Self()
		embed.SetFooter("Powered by ticketsbot.net", self.AvatarUrl(256))
	}

	// Explicitly ignore error to fix 403 (Cannot send messages to this user)
	_, err = ctx.Worker.CreateMessageEmbed(dmChannel.Id, embed)
	if err == nil {
		ctx.ReactWithCheck()
	} else {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", translations.MessageHelpDMFailed)
	}
}

func getPrefix(guildId uint64) (prefix string) {
	var err error
	prefix, err = dbclient.Client.Prefix.Get(guildId)
	if err != nil {
		sentry.Error(err)
	}

	if prefix == "" {
		prefix = utils.DEFAULT_PREFIX
	}

	return
}
