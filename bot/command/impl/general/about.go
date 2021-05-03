package general

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"time"
)

type AboutCommand struct {
}

func (AboutCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "about",
		Description:      translations.HelpAbout,
		PermissionLevel:  permission.Everyone,
		Category:         command.General,
		MainBotOnly:      true,
		DefaultEphemeral: true,
	}
}

func (c AboutCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AboutCommand) Execute(ctx registry.CommandContext) {
	time.Sleep(time.Second * 6)
	ctx.Reply(utils.Green, "About", translations.MessageAbout)
}
