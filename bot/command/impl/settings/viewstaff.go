package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
)

type ViewStaffCommand struct {
}

func (ViewStaffCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "viewstaff",
		Description:     i18n.HelpViewStaff,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
	}
}

func (c ViewStaffCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ViewStaffCommand) Execute(ctx registry.CommandContext) {
	embed, _ := logic.BuildViewStaffMessage(ctx.GuildId(), ctx.Worker(), 0, ctx.ToErrorContext())

	msg, err := ctx.Worker().CreateMessageEmbed(ctx.ChannelId(), embed)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		if err := ctx.Worker().CreateReaction(ctx.ChannelId(), msg.Id, "◀️"); err != nil {
			ctx.HandleError(err)
		}

		if err := ctx.Worker().CreateReaction(ctx.ChannelId(), msg.Id, "▶️"); err != nil {
			ctx.HandleError(err)
		}

		utils.DeleteAfter(ctx.Worker(), ctx.ChannelId(), msg.Id, 60)
	}

	redis.SetPage(redis.Client, msg.Id, 0)
}
