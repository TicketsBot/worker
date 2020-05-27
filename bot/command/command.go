package command

import (
	"github.com/TicketsBot/common/permission"
)

type Command interface {
	Name() string
	Description() string
	Aliases() []string
	PermissionLevel() permission.PermissionLevel
	Execute(ctx CommandContext)
	Parent() interface{}
	Children() []Command
	PremiumOnly() bool
	Category() Category
	AdminOnly() bool
	HelperOnly() bool
}

