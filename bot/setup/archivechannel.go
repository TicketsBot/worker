package setup

import (
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
	"strings"
)

type ArchiveChannelStage struct {
}

func (ArchiveChannelStage) State() State {
	return ArchiveChannel
}

func (ArchiveChannelStage) Prompt() string {
	return "Please specify you wish ticket logs to be sent to after tickets have been closed" +
		"\nExample: `#logs`"
}

func (ArchiveChannelStage) Default() string {
	return ""
}

func (ArchiveChannelStage) Process(worker *worker.Context, msg message.Message) {
	guild, err := worker.GetGuild(msg.GuildId); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildId,
			User:    msg.Author.Id,
			Channel: msg.ChannelId,
			Shard:   worker.ShardId,
		})
		return
	}

	var archiveChannelId uint64

	// Prefer channel mention
	mentions := msg.ChannelMentions()
	if len(mentions) > 0 {
		archiveChannelId = mentions[0]

		// Verify that the channel exists
		exists := false
		for _, guildChannel := range guild.Channels {
			if guildChannel.Id == archiveChannelId {
				exists = true
				break
			}
		}

		if !exists {
			utils.SendEmbed(worker, msg.ChannelId, utils.Red, "Error", "Invalid channel, disabling archiving", nil, 15, true)
			utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
			return
		}
	} else {
		// Try to match channel name
		split := strings.Split(msg.Content, " ")
		name := split[0]

		// Get channels from discord
		channels, err := worker.GetGuildChannels(msg.GuildId); if err != nil {
			utils.SendEmbed(worker, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()), nil, 15, true)
			return
		}

		found := false
		for _, channel := range channels {
			if channel.Name == name {
				found = true
				archiveChannelId = channel.Id
				break
			}
		}

		if !found {
			utils.SendEmbed(worker, msg.ChannelId, utils.Red, "Error", "Invalid channel, disabling archiving", nil, 15, true)
			utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
			return
		}
	}

	if err := dbclient.Client.ArchiveChannel.Set(msg.GuildId, archiveChannelId); err == nil {
		utils.ReactWithCheck(worker, msg.ChannelId, msg.Id)
	} else {
		utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
		sentry.Error(err)
	}
}
