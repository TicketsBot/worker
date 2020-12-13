package impl

import (
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/impl/admin"
	"github.com/TicketsBot/worker/bot/command/impl/general"
	"github.com/TicketsBot/worker/bot/command/impl/settings"
	"github.com/TicketsBot/worker/bot/command/impl/tags"
	"github.com/TicketsBot/worker/bot/command/impl/tickets"
)

var Commands = map[string]command.Command{
	"help": HelpCommand{},

	"admin":            admin.AdminCommand{},
	"registercommands": RegisterCommandsCommand{},

	"about": general.AboutCommand{},
	"vote": general.VoteCommand{},

	"addadmin": settings.AddAdminCommand{},
	//settings.AddSupportCommand{},
	"blacklist": settings.BlacklistCommand{},
	"cancel": settings.CancelCommand{},
	"language": settings.LanguageCommand{},
	"panel": settings.PanelCommand{},
	"premium": settings.PremiumCommand{},
	//settings.RemoveAdminCommand{},
	//settings.RemoveSupportCommand{},
	//setup.SetupCommand{},
	"viewstaff": settings.ViewStaffCommand{},

	"sync": settings.SyncCommand{},
	//statistics.StatsCommand{},

	//tags.ManageTagsCommand{},
	"tag": tags.TagCommand{},

	"add": tickets.AddCommand{},
	"claim": tickets.ClaimCommand{},
	"close": tickets.CloseCommand{},
	"open": tickets.OpenCommand{},
	//"remove": tickets.RemoveCommand{},
	"rename": tickets.RenameCommand{},
	"transfer": tickets.TransferCommand{},
	"unclaim": tickets.UnclaimCommand{},
}