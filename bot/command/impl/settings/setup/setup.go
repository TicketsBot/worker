package setup

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"golang.org/x/sync/errgroup"
)

type SetupCommand struct {
}

func (SetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "setup",
		Description:     translations.HelpSetup,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Children: []registry.Command{
			EasySetupCommand{},
			AutoSetupCommand{},
			PrefixSetupCommand{},
			WelcomeMessageSetupCommand{},
			LimitSetupCommand{},
			TranscriptsSetupCommand{},
			CategorySetupCommand{},
		},
	}
}

func (c SetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c SetupCommand) Execute(ctx registry.CommandContext) {
	ctx.ReplyWithFieldsPermanent(utils.Green, "Setup", translations.SetupChoose, c.buildFields(ctx))
}

func (SetupCommand) buildFields(ctx registry.CommandContext) []embed.EmbedField {
	fields := make([]embed.EmbedField, 9)

	group, _ := errgroup.WithContext(context.Background())

	group.Go(getFieldFunc(ctx, fields, 0, "t!setup ez", translations.SetupEasyDescription, true))
	group.Go(getFieldFunc(ctx, fields, 1, "t!setup auto", translations.SetupAutoDescription, true))
	group.Go(getFieldFunc(ctx, fields, 2, "Dashboard", translations.SetupDashboardDescription, true))
	group.Go(getFieldFunc(ctx, fields, 3, "t!setup prefix", translations.SetupPrefixDescription, true))
	group.Go(getFieldFunc(ctx, fields, 4, "t!setup limit", translations.SetupLimitDescription, true))
	group.Go(getFieldFunc(ctx, fields, 5, "t!setup welcomemessage", translations.SetupWelcomeMessageDescription, false))
	group.Go(getFieldFunc(ctx, fields, 6, "t!setup transcripts", translations.SetupTranscriptsDescription, true))
	group.Go(getFieldFunc(ctx, fields, 7, "t!setup category", translations.SetupCategoryDescription, true))
	group.Go(getFieldFunc(ctx, fields, 8, "Reaction Panels", translations.SetupReactionPanelsDescription, false, ctx.GuildId))

	// should never happen
	if err := group.Wait(); err != nil {
		sentry.Error(err)
		return nil
	}

	return fields
}

func newFieldFromTranslation(ctx registry.CommandContext, name string, value translations.MessageId, inline bool, format ...interface{}) embed.EmbedField {
	return embed.EmbedField{
		Name:   name,
		Value:  i18n.GetMessageFromGuild(ctx.GuildId(), value, format...),
		Inline: inline,
	}
}

func getFieldFunc(ctx registry.CommandContext, fields []embed.EmbedField, index int, name string, value translations.MessageId, inline bool, format ...interface{}) func() error {
	return func() error {
		fields[index] = newFieldFromTranslation(ctx, name, value, inline, format...)
		return nil
	}
}
