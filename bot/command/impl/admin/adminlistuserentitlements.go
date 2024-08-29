package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type AdminListUserEntitlementsCommand struct {
}

func (AdminListUserEntitlementsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "list-user-entitlements",
		Description:     i18n.HelpAdmin,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User to fetch entitlements for", interaction.OptionTypeUser, i18n.MessageInvalidArgument),
		),
		Timeout: time.Second * 15,
	}
}

func (c AdminListUserEntitlementsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminListUserEntitlementsCommand) Execute(ctx registry.CommandContext, userId uint64) {
	// List entitlements that have expired in the past 30 days
	entitlements, err := dbclient.Client.Entitlements.ListUserSubscriptions(ctx, userId, time.Hour*24*30)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	embed := embed.NewEmbed().
		SetTitle("Entitlements").
		SetColor(ctx.GetColour(customisation.Blue))

	if len(entitlements) == 0 {
		embed.SetDescription("No entitlements found")
	}

	for i, entitlement := range entitlements {
		if i >= 25 {
			embed.SetDescription("Too many entitlements to display")
			break
		}

		value := fmt.Sprintf(
			"**Tier:** %s\n**Source:** %s\n**Expires:** <t:%d>\n**SKU ID:** %s\n**SKU Priority:** %d",
			entitlement.Tier,
			entitlement.Source,
			entitlement.ExpiresAt.Unix(),
			entitlement.SkuId.String(),
			entitlement.SkuPriority,
		)

		embed.AddField(entitlement.SkuLabel, value, false)
	}

	ctx.ReplyWithEmbed(embed)
}
