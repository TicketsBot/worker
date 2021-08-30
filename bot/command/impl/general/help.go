package general

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/elliotchance/orderedmap"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type HelpCommand struct {
	Registry registry.Registry
}

func (HelpCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "help",
		Description:      i18n.HelpHelp,
		Aliases:          []string{"h"},
		PermissionLevel:  permission.Everyone,
		Category:         command.General,
		DefaultEphemeral: true,
	}
}

func (c HelpCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c HelpCommand) Execute(ctx registry.CommandContext) {
	commandCategories := orderedmap.NewOrderedMap()

	// initialise map with the correct order of categories
	for _, category := range command.Categories {
		commandCategories.Set(category, nil)
	}

	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	for _, cmd := range c.Registry {
		// check bot admin / helper only commands
		if (cmd.Properties().AdminOnly && !utils.IsBotAdmin(ctx.UserId())) || (cmd.Properties().HelperOnly && !utils.IsBotHelper(ctx.UserId())) {
			continue
		}

		// check whitelabel hidden cmds
		if cmd.Properties().MainBotOnly && ctx.Worker().IsWhitelabel {
			continue
		}

		if permLevel >= cmd.Properties().PermissionLevel { // only send commands the user has permissions for
			var current []registry.Command
			if commands, ok := commandCategories.Get(cmd.Properties().Category); ok {
				if commands == nil {
					current = make([]registry.Command, 0)
				} else {
					current = commands.([]registry.Command)
				}
			}
			current = append(current, cmd)

			commandCategories.Set(cmd.Properties().Category, current)
		}
	}

	embed := embed.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Help")

	for _, category := range commandCategories.Keys() {
		var commands []registry.Command
		if retrieved, ok := commandCategories.Get(category.(command.Category)); ok {
			if retrieved == nil {
				commands = make([]registry.Command, 0)
			} else {
				commands = retrieved.([]registry.Command)
			}
		}

		if len(commands) > 0 {
			formatted := make([]string, 0)
			for _, cmd := range commands {
				formatted = append(formatted, registry.FormatHelp(cmd, ctx.GuildId()))
			}

			embed.AddField(string(category.(command.Category)), strings.Join(formatted, "\n"), false)
		}
	}

	if ctx.PremiumTier() == premium.None {
		embed.SetFooter("Powered by ticketsbot.net", "https://ticketsbot.net/assets/img/logo.png")
	}

	// Explicitly ignore error to fix 403 (Cannot send messages to this user)
	_, _ = ctx.ReplyWith(registry.NewEphemeralEmbedMessageResponse(embed))
}
