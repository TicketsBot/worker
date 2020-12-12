package command

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database/translations"
)

type Properties struct {
	Name            string
	Description     database.MessageId
	Aliases         []string
	PermissionLevel permission.PermissionLevel
	Children        []Command
	PremiumOnly     bool
	Category        Category
	AdminOnly       bool
	HelperOnly      bool
	Arguments       []Argument
}
