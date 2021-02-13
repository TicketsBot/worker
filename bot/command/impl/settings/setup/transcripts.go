package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest/request"
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
	if _, err := ctx.Worker().GetChannel(channelId); err != nil {
		if restError, ok := err.(request.RestError); ok && restError.IsClientError() {
			ctx.Reply(utils.Red, "Error", translations.SetupTranscriptsInvalid, ctx.ChannelId)
			ctx.Reject()
		} else {
			ctx.HandleError(err)
		}

		return
	}
	if err := dbclient.Client.ArchiveChannel.Set(ctx.GuildId(), channelId); err == nil {
		ctx.Accept()
		ctx.Reply(utils.Green, "Setup", translations.SetupTranscriptsComplete, channelId)
	} else {
		ctx.HandleError(err)
	}
}
