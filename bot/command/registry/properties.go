package registry

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type Properties struct {
	Name             string
	Description      i18n.MessageId
	Type             interaction.ApplicationCommandType
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
	Timeout          time.Duration

	SetupFunc func()
}
