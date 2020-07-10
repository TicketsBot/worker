package setup

import (
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
	"strings"
)

type ChannelCategoryStage struct {
}

func (ChannelCategoryStage) State() State {
	return ChannelCategory
}

func (ChannelCategoryStage) Prompt() translations.MessageId {
	return translations.SetupChannelCategory
}

func (ChannelCategoryStage) Default() string {
	return ""
}

func (ChannelCategoryStage) Process(worker *worker.Context, msg message.Message) {
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   msg.GuildId,
		User:    msg.Author.Id,
		Channel: msg.ChannelId,
		Shard:   worker.ShardId,
	}

	name := msg.Content

	guild, err := worker.GetGuild(msg.GuildId); if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	var categoryId uint64
	for _, ch := range guild.Channels {
		if ch.Type == channel.ChannelTypeGuildCategory && strings.ToLower(ch.Name) == strings.ToLower(name) {
			categoryId = ch.Id
			break
		}
	}

	if categoryId == 0 {
		// Attempt to create categoryName
		data := rest.CreateChannelData{
			Name: name,
			Type: channel.ChannelTypeGuildCategory,
		}

		category, err := worker.CreateGuildChannel(guild.Id, data); if err != nil {
			// Likely no permission, default to having no category
			utils.SendEmbed(worker, msg.ChannelId, msg.GuildId, utils.Red, "Error", translations.MessageInvalidCategory, nil, 15, true)
			return
		}

		categoryId = category.Id

		utils.SendEmbed(worker, msg.ChannelId, msg.GuildId, utils.Red, "Error", translations.MessageCreatedCategory, nil, 15, true, category.Name)
	}

	if err := dbclient.Client.ChannelCategory.Set(msg.GuildId, categoryId); err == nil {
		utils.ReactWithCheck(worker, msg.ChannelId, msg.Id)
	} else {
		utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
		sentry.Error(err)
	}
}
