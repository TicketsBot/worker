package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/interaction"
)

type TranscriptsSetupCommand struct{}

func (TranscriptsSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "transcripts",
		Description:     translations.HelpSetup,
		Aliases:         []string{"transcript", "archives", "archive"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("channel", "The channel that ticket transcripts should be sent to", interaction.OptionTypeChannel, translations.SetupTranscriptsInvalid),
		),
	}
}

func (c TranscriptsSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (TranscriptsSetupCommand) Execute(ctx command.CommandContext, channelId uint64) {
	channels, err := ctx.Worker().GetGuildChannels(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify that the channel exists
	exists := false
	for _, ch := range channels {
		if ch.Id == channelId && ch.Type == channel.ChannelTypeGuildText {
			exists = true
			break
		}
	}

	if !exists {
		ctx.Reply(utils.Red, "Error", translations.SetupTranscriptsInvalid, ctx.ChannelId)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.ArchiveChannel.Set(ctx.GuildId(), channelId); err == nil {
		ctx.Accept()
		ctx.Reply(utils.Green, "Setup", translations.SetupTranscriptsComplete, channelId)
	} else {
		ctx.HandleError(err)
	}
}