package impl

import (
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/impl/admin"
	"github.com/TicketsBot/worker/bot/command/impl/general"
	"github.com/TicketsBot/worker/bot/command/impl/settings"
	"github.com/TicketsBot/worker/bot/command/impl/statistics"
	"github.com/TicketsBot/worker/bot/command/impl/tags"
	"github.com/TicketsBot/worker/bot/command/impl/tickets"
)

var Commands = []command.Command{
	HelpCommand{},

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
	settings.SetupCommand{},
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
	tickets.UnclaimCommand{},
}