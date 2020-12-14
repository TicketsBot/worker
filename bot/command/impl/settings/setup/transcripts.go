package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
)

type TranscriptsSetupCommand struct{}

func (TranscriptsSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "transcripts",
		Description:     translations.HelpSetup,
		Aliases:         []string{"transcript", "archives", "archive"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (TranscriptsSetupCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply(utils.Red, "Setup", translations.SetupTranscriptsInvalid, ctx.ChannelId)
		ctx.ReactWithCross()
		return
	}

	var transcriptsChannelId uint64

	// Prefer channel mention
	// TODO: Remove code repetition
	mentions := ctx.ChannelMentions()
	if len(mentions) > 0 {
		transcriptsChannelId = mentions[0]

		// get guild object
		guild, err := ctx.Guild()
		if err != nil {
			ctx.HandleError(err)
			return
		}

		// Verify that the channel exists
		exists := false
		for _, guildChannel := range guild.Channels {
			if guildChannel.Id == transcriptsChannelId {
				exists = true
				break
			}
		}

		if !exists {
			ctx.Reply(utils.Red, "Error", translations.SetupTranscriptsInvalid, ctx.ChannelId)
			ctx.ReactWithCross()
			return
		}
	} else { // Try to match channel name
		name := ctx.Args[0]

		// Get channels from discord
		channels, err := ctx.Worker.GetGuildChannels(ctx.GuildId); if err != nil {
			ctx.HandleError(err)
			return
		}

		found := false
		for _, channel := range channels {
			if channel.Name == name {
				found = true
				transcriptsChannelId = channel.Id
				break
			}
		}

		if !found {
			ctx.Reply(utils.Red, "Error", translations.SetupTranscriptsInvalid, ctx.ChannelId)
			ctx.ReactWithCross()
			return
		}
	}

	if err := dbclient.Client.ArchiveChannel.Set(ctx.GuildId, transcriptsChannelId); err == nil {
		ctx.ReactWithCheck()
		ctx.Reply(utils.Green, "Setup", translations.SetupTranscriptsComplete, transcriptsChannelId)
	} else {
		ctx.HandleError(err)
	}
}
