package handlers

import (
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction/component"
	"regexp"
	"strconv"
	"strings"
)

type ViewSurveyHandler struct{}

func (h *ViewSurveyHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "view-survey-")
	})
}

func (h *ViewSurveyHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

var viewSurveypattern = regexp.MustCompile(`view-survey-(\d+)-(\d+)`)

func (h *ViewSurveyHandler) Execute(ctx *context.ButtonContext) {
	groups := viewSurveypattern.FindStringSubmatch(ctx.InteractionData.CustomId)
	if len(groups) != 3 {
		return
	}

	// Error may occur if guild ID in custom ID > max u64 size
	guildId, err := strconv.ParseUint(groups[1], 10, 64)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ticketId, err := strconv.Atoi(groups[2])
	if err != nil {
		ctx.HandleError(err)
		return
	}

	premiumTier, err := utils.PremiumClient.GetTierByGuildId(guildId, true, ctx.Worker().Token, ctx.Worker().RateLimiter)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if premiumTier == premium.None {
		ctx.ReplyRaw(customisation.Red, "Error", "Premium is required for exit surveys.") // TODO: i18n
		return
	}

	// Get ticket
	ticket, err := dbclient.Client.Tickets.Get(ticketId, guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.UserId != ctx.InteractionUser().Id || ticket.GuildId != guildId || ticket.Id != ticketId {
		return
	}

	response, err := dbclient.Client.ExitSurveyResponses.GetResponses(ticket.GuildId, ticket.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if len(response.Responses) == 0 {
		ctx.ReplyRaw(customisation.Red, "Error", "No survey response has been recorded for this ticket.") // TODO: i18n
		return
	}

	opener, err := ctx.Worker().GetUser(ticket.UserId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	e := embed.NewEmbed().
		SetTitle("Exit Survey"). // TODO: i18n
		SetAuthor(opener.Username, "", opener.AvatarUrl(256)).
		SetColor(ctx.GetColour(customisation.Green))

	for _, answer := range response.Responses {
		var title string
		if answer.Question == nil {
			title = "Unknown Question"
		} else {
			title = *answer.Question
		}

		e.AddField(title, answer.Response, false)
	}

	var buttons []component.Component
	buttons = append(buttons, logic.TranscriptLinkElement(ticket.HasTranscript)(ctx, ticket)...)
	buttons = append(buttons, logic.ThreadLinkElement(ticket.ChannelId != nil && ticket.IsThread)(ctx, ticket)...)

	if len(buttons) > 0 {
		ctx.ReplyWithEmbedAndComponents(e, utils.Slice(component.BuildActionRow(buttons...)))
	} else {
		ctx.ReplyWithEmbed(e)
	}
}
