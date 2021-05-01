package registry

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
)

type Properties struct {
	Name             string
	Description      database.MessageId
	Aliases          []string
	PermissionLevel  permission.PermissionLevel
	Children         []Command // TODO: Map
	PremiumOnly      bool
	Category         command.Category
	AdminOnly        bool
	HelperOnly       bool
	InteractionOnly  bool
	MessageOnly      bool
	MainBotOnly      bool
	Arguments        []command.Argument
	DefaultEphemeral bool
}
