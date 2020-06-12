package command

import "github.com/TicketsBot/common/permission"

type Properties struct {
	Name            string
	Description     string
	Aliases         []string
	PermissionLevel permission.PermissionLevel
	Children        []Command
	PremiumOnly     bool
	Category        Category
	AdminOnly       bool
	HelperOnly      bool
}
