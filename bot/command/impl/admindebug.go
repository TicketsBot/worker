package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strconv"
)

type AdminDebugCommand struct {
}

func (AdminDebugCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "debug",
		Description:     "Provides debugging information",
		PermissionLevel: permission.Everyone,
		Parent:          AdminCommand{},
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (AdminDebugCommand) Execute(ctx command.CommandContext) {
	// Get ticket category
	categoryId, err := dbclient.Client.ChannelCategory.Get(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// get guild channels
	channels, err := ctx.Worker.GetGuildChannels(ctx.GuildId); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	var categoryName string
	for _, channel := range channels {
		if channel.Id == categoryId { // Don't need to compare channel types
			categoryName = channel.Name
		}
	}

	if categoryName == "" {
		categoryName = "None"
	}

	// get guild object
	guild, err := ctx.Guild(); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Get owner
	invalidOwner := false
	owner, err := ctx.Worker.GetGuildMember(ctx.GuildId, guild.OwnerId); if err != nil {
		invalidOwner = true
	}

	var ownerFormatted string
	if invalidOwner {
		ownerFormatted = strconv.FormatUint(guild.OwnerId, 10)
	} else {
		ownerFormatted = fmt.Sprintf("%s#%s", owner.User.Username, utils.PadDiscriminator(owner.User.Discriminator))
	}

	// Get archive channel
	//archiveChannelChan := make(chan int64)
	//go database.GetArchiveChannel()

	embed := embed.NewEmbed().
		SetTitle("Admin").
		SetColor(int(utils.Green)).

		AddField("Shard", strconv.Itoa(ctx.Worker.ShardId), true).
		AddBlankField(false).

		AddField("Ticket Category", categoryName, true).
		AddField("Owner", ownerFormatted, true)

	msg, err := ctx.Worker.CreateMessageEmbed(ctx.ChannelId, embed); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	utils.DeleteAfter(utils.SentMessage{Worker: ctx.Worker, Message: &msg}, 30)
}
