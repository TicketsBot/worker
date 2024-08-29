package admin

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/google/uuid"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
	"time"
)

type AdminGenPremiumCommand struct {
}

func (c AdminGenPremiumCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "genpremium",
		Description:     i18n.HelpAdminGenPremium,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		Arguments: command.Arguments(
			command.NewRequiredAutocompleteableArgument("sku", "SKU for the key to grant", interaction.OptionTypeString, i18n.MessageInvalidArgument, c.AutoCompleteHandler),
			command.NewRequiredArgument("length", "Length in days of the key", interaction.OptionTypeInteger, i18n.MessageInvalidArgument),
			command.NewOptionalArgument("amount", "Amount of keys to generate", interaction.OptionTypeInteger, i18n.MessageInvalidArgument),
		),
		Timeout: time.Second * 10,
	}
}

func (c AdminGenPremiumCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminGenPremiumCommand) Execute(ctx registry.CommandContext, skuIdRaw string, length int, amountRaw *int) {
	amount := 1
	if amountRaw != nil {
		amount = *amountRaw
	}

	skuId, err := uuid.Parse(skuIdRaw)
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Admin), "Invalid SKU")
		return
	}

	tx, err := dbclient.Client.BeginTx(ctx)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		tx.Rollback(ctx)
	}()

	sku, err := dbclient.Client.SubscriptionSkus.GetSku(ctx, tx, skuId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if sku == nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Admin), "Invalid SKU")
		return
	}

	keys := make([]string, 0, amount)
	for i := 0; i < amount; i++ {
		key := uuid.New()

		if err := dbclient.Client.PremiumKeys.Create(ctx, key, time.Hour*24*time.Duration(length), skuId); err != nil {
			ctx.HandleError(err)
			return
		}

		keys = append(keys, key.String())
	}

	dmChannel, err := ctx.Worker().CreateDM(ctx.UserId())
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Admin), err.Error())
		return
	}

	content := "```\n"
	for _, key := range keys {
		content += fmt.Sprintf("%s\n", key)
	}
	content = strings.TrimSuffix(content, "\n")
	content += "```"

	_, err = ctx.Worker().CreateMessage(dmChannel.Id, content)
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Admin), err.Error())
		return
	}

	ctx.ReplyPlainPermanent("Check your DMs")
}

func (AdminGenPremiumCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	skus, err := dbclient.Client.SubscriptionSkus.Search(ctx, value, 10)
	if err != nil {
		sentry.Error(err)
		return nil
	}

	choices := make([]interaction.ApplicationCommandOptionChoice, len(skus))
	for i, sku := range skus {
		choices[i] = interaction.ApplicationCommandOptionChoice{
			Name:  sku.Label,
			Value: sku.Id.String(),
		}
	}

	return choices
}
