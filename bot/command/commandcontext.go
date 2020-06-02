package command

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"strconv"
	"strings"
)

type CommandContext struct {
	Worker *worker.Context
	message.Message
	Root                string
	Args                []string
	PremiumTier         premium.PremiumTier
	ShouldReact         bool
	IsFromPanel         bool
	UserPermissionLevel permission.PermissionLevel
}

func (ctx *CommandContext) Guild() (guild.Guild, error) {
	return ctx.Worker.GetGuild(ctx.GuildId)
}

func (ctx *CommandContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:       ctx.GuildId,
		User:        ctx.Author.Id,
		Channel:     ctx.ChannelId,
		Command:     ctx.Root + " " + strings.Join(ctx.Args, " "),
	}
}

func (ctx *CommandContext) SendEmbed(colour utils.Colour, title, content string, fields ...embed.EmbedField) {
	utils.SendEmbed(ctx.Worker, ctx.ChannelId, colour, title, content, fields, 30, ctx.PremiumTier > premium.None)
}

func (ctx *CommandContext) SendEmbedNoDelete(colour utils.Colour, title, content string, fields ...embed.EmbedField) {
	utils.SendEmbed(ctx.Worker, ctx.ChannelId, colour, title, content, fields, 0, ctx.PremiumTier > premium.None)
}

func (ctx *CommandContext) SendMessage(content string) {
	msg, err := ctx.Worker.CreateMessage(ctx.ChannelId, content)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		utils.DeleteAfter(utils.SentMessage{Worker: ctx.Worker, Message: &msg}, 60)
	}
}

func (ctx *CommandContext) ReactWithCheck() {
	if ctx.ShouldReact {
		utils.ReactWithCheck(ctx.Worker, ctx.ChannelId, ctx.Id)
	}
}

func (ctx *CommandContext) ReactWithCross() {
	if ctx.ShouldReact {
		utils.ReactWithCross(ctx.Worker, ctx.ChannelId, ctx.Id)
	}
}

func (ctx *CommandContext) GetPermissionLevel() permission.PermissionLevel {
	return permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker), ctx.Member, ctx.GuildId)
}

func (ctx *CommandContext) GetChannelFromArgs() uint64 {
	mentions := ctx.ChannelMentions()
	if len(mentions) > 0 {
		return mentions[0]
	}

	for _, arg := range ctx.Args {
		if parsed, err := strconv.ParseUint(arg, 10, 64); err == nil {
			return parsed
		}
	}

	return 0
}

func (ctx *CommandContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())
	ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
}
