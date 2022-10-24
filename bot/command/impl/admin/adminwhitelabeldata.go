package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type AdminWhitelabelDataCommand struct {
}

func (AdminWhitelabelDataCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "whitelabel-data",
		Description:     i18n.HelpAdmin,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user_id", "ID of the user who has the whitelabel subscription", interaction.OptionTypeUser, i18n.MessageInvalidArgument),
		),
	}
}

func (c AdminWhitelabelDataCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminWhitelabelDataCommand) Execute(ctx registry.CommandContext, userId uint64) {
	tier, err := utils.PremiumClient.GetTierByUser(userId, false)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if tier < premium.Whitelabel {
		ctx.ReplyRaw(customisation.Red, "Subscription not found", fmt.Sprintf("User does not have a whitelabel subscription (%s)", tier.String()))
		return
	}

	data, err := dbclient.Client.Whitelabel.GetByUserId(userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	var botIdFormatted = "Bot not found"
	var publicKey = "Not set"
	if data.BotId != 0 {
		botIdFormatted = fmt.Sprintf("%d (<@%d>)", data.BotId, data.BotId)

		publicKey, err = dbclient.Client.WhitelabelKeys.Get(data.BotId)
		if err != nil {
			ctx.HandleError(err)
			return
		}
	}

	errors, err := dbclient.Client.WhitelabelErrors.GetRecent(userId, 3)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	var errorsFormatted string
	if len(errors) == 0 {
		errorsFormatted = "No errors found"
	} else {
		strs := make([]string, len(errors))
		for i, botError := range errors {
			strs[i] = fmt.Sprintf("[<t:%d:f>] `%s`", botError.Time.Unix(), botError.Message)
		}

		errorsFormatted = strings.Join(strs, "\n")
	}

	fields := []embed.EmbedField{
		utils.EmbedFieldRaw("Subscription Tier", tier.String(), true),
		utils.EmbedFieldRaw("Bot ID", botIdFormatted, true),
		utils.EmbedFieldRaw("Public Key", publicKey, false),
		utils.EmbedFieldRaw("Last 3 Errors", errorsFormatted, false),
	}

	ctx.ReplyWithEmbed(utils.BuildEmbedRaw(ctx.GetColour(customisation.Green), "Whitelabel", "", fields, ctx.PremiumTier()))
}
