package impl

import "github.com/TicketsBot/worker/bot/command"

var Commands = []command.Command{
	AboutCommand{},
	AddCommand{},
	AddAdminCommand{},
	AddSupportCommand{},
	AdminCommand{},
	BlacklistCommand{},
	CancelCommand{},
	ClaimCommand{},
	CloseCommand{},
	HelpCommand{},
	ManageTagsCommand{},
	OpenCommand{},
	PanelCommand{},
	PremiumCommand{},
	RemoveCommand{},
	RemoveAdminCommand{},
	RemoveSupportCommand{},
	RenameCommand{},
	SetupCommand{},
	StatsCommand{},
	SyncCommand{},
	TagCommand{},
	TransferCommand{},
	UnclaimCommand{},
	ViewStaffCommand{},
	VoteCommand{},
}