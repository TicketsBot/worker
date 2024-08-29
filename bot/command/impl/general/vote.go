package general

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/jackc/pgx/v4"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"time"
)

type VoteCommand struct {
}

func (VoteCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "vote",
		Description:      i18n.HelpVote,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Everyone,
		Category:         command.General,
		DefaultEphemeral: true,
		MainBotOnly:      true,
		Timeout:          time.Second * 3,
	}
}

func (c VoteCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c VoteCommand) Execute(ctx registry.CommandContext) {
	var credits int
	if err := dbclient.Client.WithTx(ctx, func(tx pgx.Tx) (err error) {
		credits, err = dbclient.Client.VoteCredits.Get(ctx, tx, ctx.UserId())
		return
	}); err != nil {
		ctx.HandleError(err)
		return
	}

	if credits == 0 {
		commandIds, err := command.LoadCommandIds(ctx.Worker(), ctx.Worker().BotId)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		var commandMention string
		if id, ok := commandIds[c.Properties().Name]; ok {
			commandMention = fmt.Sprintf("</%s:%d>", c.Properties().Name, id)
		} else {
			commandMention = "`/vote`"
		}

		embed := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleVote, i18n.MessageVote, nil, commandMention)

		if _, err := ctx.ReplyWith(command.NewEphemeralEmbedMessageResponseWithComponents(embed, buildVoteComponents(ctx, false))); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		var embed *embed.Embed
		if credits == 1 {
			embed = utils.BuildEmbed(ctx, customisation.Green, i18n.TitleVote, i18n.MessageVoteWithCreditsSingular, nil, credits, credits)
		} else {
			embed = utils.BuildEmbed(ctx, customisation.Green, i18n.TitleVote, i18n.MessageVoteWithCreditsPlural, nil, credits, credits)
		}

		if _, err := ctx.ReplyWith(command.NewEphemeralEmbedMessageResponseWithComponents(embed, buildVoteComponents(ctx, true))); err != nil {
			ctx.HandleError(err)
			return
		}
	}
}

func buildVoteComponents(ctx registry.CommandContext, allowRedeem bool) []component.Component {
	voteButton := component.BuildButton(component.Button{
		Label: ctx.GetMessage(i18n.TitleVote),
		Style: component.ButtonStyleLink,
		Emoji: utils.BuildEmoji("🔗"),
		Url:   utils.Ptr("https://vote.ticketsbot.net"),
	})

	redeemButton := component.BuildButton(component.Button{
		Label:    ctx.GetMessage(i18n.MessageVoteRedeemCredits),
		CustomId: "redeem_vote_credits",
		Style:    component.ButtonStylePrimary,
		Emoji:    utils.BuildEmoji("💶"),
	})

	var actionRow component.Component
	if allowRedeem {
		actionRow = component.BuildActionRow(voteButton, redeemButton)
	} else {
		actionRow = component.BuildActionRow(voteButton)
	}

	return []component.Component{actionRow}
}
