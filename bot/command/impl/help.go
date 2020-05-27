package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/elliotchance/orderedmap"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type HelpCommand struct {
}

func (HelpCommand) Name() string {
	return "help"
}

func (HelpCommand) Description() string {
	return "Shows you a list of commands"
}

func (HelpCommand) Aliases() []string {
	return []string{"h"}
}

func (HelpCommand) PermissionLevel() permission.PermissionLevel {
	return permission.Everyone
}

func (HelpCommand) Execute(ctx command.CommandContext) {
	commandCategories := orderedmap.NewOrderedMap()

	// initialise map with the correct order of categories
	for _, category := range command.Categories {
		commandCategories.Set(category, nil)
	}

	for _, cmd := range Commands {
		// check bot admin / helper only commands
		if (cmd.AdminOnly() && !utils.IsBotAdmin(ctx.Author.Id)) || (cmd.HelperOnly() && !utils.IsBotHelper(ctx.Author.Id)) {
			continue
		}

		if ctx.UserPermissionLevel >= cmd.PermissionLevel() { // only send commands the user has permissions for
			var current []command.Command
			if commands, ok := commandCategories.Get(cmd.Category()); ok {
				if commands == nil {
					current = make([]command.Command, 0)
				} else {
					current = commands.([]command.Command)
				}
			}
			current = append(current, cmd)

			commandCategories.Set(cmd.Category(), current)
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
			for _, command := range commands {
				formatted = append(formatted, formatHelp(command, prefix))
			}

			embed.AddField(string(category.(command.Category)), strings.Join(formatted, "\n"), false)
		}
	}

	dmChannel, err := ctx.Worker.CreateDM(ctx.Author.Id); if err != nil {
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
		ctx.SendEmbed(utils.Red, "Error", "I couldn't send you a direct message: make sure your privacy settings aren't too high")
	}
}

func formatHelp(c command.Command, prefix string) string {
	return fmt.Sprintf("**%s%s**: %s", prefix, c.Name(), c.Description())
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

func (HelpCommand) Parent() interface{} {
	return nil
}

func (HelpCommand) Children() []command.Command {
	return nil
}

func (HelpCommand) PremiumOnly() bool {
	return false
}

func (HelpCommand) Category() command.Category {
	return command.General
}

func (HelpCommand) AdminOnly() bool {
	return false
}

func (HelpCommand) HelperOnly() bool {
	return false
}

