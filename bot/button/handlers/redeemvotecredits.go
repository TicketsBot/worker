package handlers

import (
	"errors"
	"github.com/TicketsBot/common/model"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/TicketsBot/worker/i18n"
	"github.com/jackc/pgx/v4"
	"github.com/rxdn/gdl/objects/interaction/component"
	"time"
)

type RedeemVoteCreditsHandler struct{}

func (h *RedeemVoteCreditsHandler) Matcher() matcher.Matcher {
	return &matcher.SimpleMatcher{
		CustomId: "redeem_vote_credits",
	}
}

func (h *RedeemVoteCreditsHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:           registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
		PermissionLevel: permission.Support,
		Timeout:         time.Second * 5,
	}
}

var errNoCredits = errors.New("no credits")

func (h *RedeemVoteCreditsHandler) Execute(ctx *context.ButtonContext) {
	var credits int
	if err := dbclient.Client.WithTx(ctx, func(tx pgx.Tx) error {
		var err error
		credits, err = dbclient.Client.VoteCredits.Get(ctx, tx, ctx.UserId())
		if err != nil {
			return err
		}

		if credits <= 0 {
			ctx.EditWithComponents(customisation.Red, i18n.Error, i18n.MessageVoteNoCredits, make([]component.Component, 0))
			return errNoCredits
		}

		if err := dbclient.Client.VoteCredits.Delete(ctx, tx, ctx.UserId()); err != nil {
			return err
		}

		if err := dbclient.Client.Entitlements.IncreaseExpiry(
			ctx,
			tx,
			utils.Ptr(ctx.GuildId()),
			utils.Ptr(ctx.UserId()),
			config.Conf.VoteSkuId,
			model.EntitlementSourceVoting,
			time.Hour*24*time.Duration(credits),
		); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, errNoCredits) {
			return
		}

		ctx.HandleError(err)
		return
	}

	// TODO: dbclient.Client.Panels.EnableAll?

	if err := utils.PremiumClient.DeleteCachedTier(ctx, ctx.GuildId()); err != nil {
		ctx.HandleError(err)
		return
	}

	if credits == 1 {
		ctx.EditWithComponents(customisation.Green, i18n.Success, i18n.MessageVoteRedeemSuccessSingular, make([]component.Component, 0), credits)
	} else {
		ctx.EditWithComponents(customisation.Green, i18n.Success, i18n.MessageVoteRedeemSuccessPlural, make([]component.Component, 0), credits)
	}
}
