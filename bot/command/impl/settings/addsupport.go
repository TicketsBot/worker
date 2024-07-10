package settings

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"time"
)

type AddSupportCommand struct{}

func (AddSupportCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "addsupport",
		Description:     i18n.HelpAddSupport,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"addsuport"},
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		InteractionOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("role", "Role to apply the support representative permission to", interaction.OptionTypeMentionable, i18n.MessageAddSupportNoMembers),
		),
		DefaultEphemeral: true,
		Timeout:          time.Second * 3,
	}
}

func (c AddSupportCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c AddSupportCommand) Execute(ctx registry.CommandContext, id uint64) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`/addsupport @Role`",
		Inline: false,
	}

	mentionableType, valid := context.DetermineMentionableType(ctx, id)
	if !valid {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageAddSupportNoMembers, utils.ToSlice(usageEmbed))
		return
	}

	var mention string
	if mentionableType == context.MentionableTypeUser {
		ctx.ReplyRaw(customisation.Red, "Error", "Users in support teams are now deprecated. Please use roles instead.")
		return

		//mention = fmt.Sprintf("<@%d>", id)
	} else if mentionableType == context.MentionableTypeRole {
		mention = fmt.Sprintf("<@&%d>", id)
	} else {
		ctx.HandleError(fmt.Errorf("unknown mentionable type: %d", mentionableType))
		return
	}

	// Send confirmation message
	e := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleAddSupport, i18n.MessageAddSupportConfirm, nil, mention)
	res := command.NewEphemeralEmbedMessageResponseWithComponents(e, utils.Slice(component.BuildActionRow(
		component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.Confirm),
			CustomId: fmt.Sprintf("addsupport-%d-%d", mentionableType, id),
			Style:    component.ButtonStylePrimary,
			Emoji:    nil,
		}),
	)))

	if _, err := ctx.ReplyWith(res); err != nil {
		ctx.HandleError(err)
	}
}
