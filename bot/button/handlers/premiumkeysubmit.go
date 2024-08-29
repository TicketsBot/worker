package handlers

import (
	"context"
	"errors"
	"github.com/TicketsBot/common/model"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	commandcontext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/google/uuid"
	"time"
)

type PremiumKeySubmitHandler struct{}

func (h *PremiumKeySubmitHandler) Matcher() matcher.Matcher {
	return matcher.NewSimpleMatcher("premium_key_modal")
}

func (h *PremiumKeySubmitHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:           registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
		PermissionLevel: permission.Admin,
		Timeout:         time.Second * 5,
	}
}

func (h *PremiumKeySubmitHandler) Execute(ctx *commandcontext.ModalContext) {
	key, ok := ctx.GetInput("key")
	if !ok {
		ctx.HandleError(errors.New("key not found")) // Infallible providing non malicious
		return
	}

	parsed, err := uuid.Parse(key)
	if err != nil {
		ctx.EditWith(customisation.Red, i18n.TitlePremium, i18n.MessageInvalidPremiumKey)
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

	length, skuId, ok, err := dbclient.Client.PremiumKeys.Delete(ctx, tx, parsed)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !ok {
		ctx.EditWith(customisation.Red, i18n.Error, i18n.MessageInvalidPremiumKey)
		return
	}

	if err := dbclient.Client.UsedKeys.Set(ctx, tx, parsed, ctx.GuildId(), ctx.UserId()); err != nil {
		ctx.HandleError(err)
		return
	}

	sku, err := dbclient.Client.SubscriptionSkus.GetSku(ctx, tx, skuId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if sku == nil {
		ctx.ReplyRaw(customisation.Red, "Error", "Unknown SKU")
		return
	}

	var guildId *uint64
	if !sku.IsGlobal {
		guildId = utils.Ptr(ctx.GuildId())
	}

	expiresAt := time.Now().Add(length)
	if _, err := dbclient.Client.Entitlements.Create(ctx, tx, guildId, utils.Ptr(ctx.UserId()), skuId, model.EntitlementSourceKey, &expiresAt); err != nil {
		ctx.HandleError(err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		ctx.HandleError(err)
		return
	}

	// Re-enable panels
	if err := dbclient.Client.Panel.EnableAll(ctx, ctx.GuildId()); err != nil {
		ctx.HandleError(err)
		return
	}

	if err := utils.PremiumClient.DeleteCachedTier(ctx, ctx.GuildId()); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.EditWith(customisation.Green, i18n.TitlePremium, i18n.MessagePremiumSuccess, int(length.Hours()/24))
}
