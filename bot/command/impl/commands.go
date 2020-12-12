package impl

import (
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/impl/settings"
)

var Commands = []command.Command{
	/*HelpCommand{},

	admin.AdminCommand{},

	general.AboutCommand{},
	general.VoteCommand{},

	settings.AddAdminCommand{},
	settings.AddSupportCommand{},
	settings.BlacklistCommand{},
	settings.CancelCommand{},
	settings.LanguageCommand{},
	settings.PanelCommand{},
	settings.PremiumCommand{},
	settings.RemoveAdminCommand{},
	settings.RemoveSupportCommand{},
	setup.SetupCommand{},
	settings.ViewStaffCommand{},

	settings.SyncCommand{},
	statistics.StatsCommand{},

	tags.ManageTagsCommand{},
	tags.TagCommand{},

	tickets.AddCommand{},
	tickets.ClaimCommand{},
	tickets.CloseCommand{},
	tickets.OpenCommand{},
	tickets.RemoveCommand{},
	tickets.RenameCommand{},
	tickets.TransferCommand{},
	tickets.UnclaimCommand{},*/

	settings.AddAdminCommand{},
}