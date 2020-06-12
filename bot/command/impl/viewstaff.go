package impl

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
)

type ViewStaffCommand struct {
}

func (ViewStaffCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "viewstaff",
		Description:     "Lists the staff members and roles",
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
	}
}

func (ViewStaffCommand) Execute(ctx command.CommandContext) {
	embed := logic.BuildViewStaffMessage(ctx.GuildId, 0, ctx.ToErrorContext())

	msg, err := ctx.Worker.CreateMessageEmbed(ctx.ChannelId, embed)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		if err := ctx.Worker.CreateReaction(ctx.ChannelId, msg.Id, "◀️"); err != nil {
			ctx.HandleError(err)
		}

		if err := ctx.Worker.CreateReaction(ctx.ChannelId, msg.Id, "▶️"); err != nil {
			ctx.HandleError(err)
		}

		utils.DeleteAfter(utils.SentMessage{Worker: ctx.Worker, Message: &msg}, 60)
	}

	redis.SetPage(redis.Client, msg.Id, 0)
}
