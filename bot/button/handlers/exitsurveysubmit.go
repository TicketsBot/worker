package handlers

import (
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"regexp"
	"strconv"
	"strings"
)

type ExitSurveySubmitHandler struct{}

func (h *ExitSurveySubmitHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "exit-survey-")
	})
}

func (h *ExitSurveySubmitHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.DMsAllowed),
	}
}

var exitSurveyPattern = regexp.MustCompile(`exit-survey-(\d+)-(\d+)`)

func (h *ExitSurveySubmitHandler) Execute(ctx *context.ModalContext) {
	groups := exitSurveyPattern.FindStringSubmatch(ctx.Interaction.Data.CustomId)
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
		ctx.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
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

	feedbackEnabled, err := dbclient.Client.FeedbackEnabled.Get(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !feedbackEnabled {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageFeedbackDisabled)
		return
	}

	if ticket.PanelId == nil {
		ctx.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	panel, err := dbclient.Client.Panel.GetById(*ticket.PanelId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if panel.GuildId != guildId || panel.PanelId != *ticket.PanelId {
		ctx.HandleError(fmt.Errorf("panel not found"))
		return
	}

	if panel.ExitSurveyFormId == nil {
		ctx.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	form, ok, err := dbclient.Client.Forms.Get(*panel.ExitSurveyFormId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !ok {
		ctx.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	formInputs, err := dbclient.Client.FormInput.GetInputs(form.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	responses := make(map[int]string)
	for _, input := range formInputs {
		value, ok := ctx.GetInput(input.CustomId)
		if ok {
			responses[input.Id] = value
		}
	}

	if err := dbclient.Client.ExitSurveyResponses.AddResponses(guildId, ticketId, *panel.ExitSurveyFormId, responses); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.EditWithRaw(customisation.Green, "Success", "Thank you for your feedback!") // TODO: i18n

	if err := addViewFeedbackButton(ctx, ticket); err != nil {
		ctx.HandleError(err)
		return
	}
}

func addViewFeedbackButton(ctx *context.ModalContext, ticket database.Ticket) error {
	// Get archive message
	settings, err := ctx.Settings()
	if err != nil {
		return err
	}

	closeMetadata, ok, err := dbclient.Client.CloseReason.Get(ticket.GuildId, ticket.Id)
	if err != nil {
		return err
	}

	var closedBy uint64
	var reason *string
	if ok {
		reason = closeMetadata.Reason

		if closeMetadata.ClosedBy != nil {
			closedBy = *closeMetadata.ClosedBy
		}
	}

	rating, ok, err := dbclient.Client.ServiceRatings.Get(ticket.GuildId, ticket.Id)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("exit survey was completed, but no rating was found (%d:%d)", ticket.GuildId, ticket.Id)
	}

	return logic.EditGuildArchiveMessageIfExists(ctx.Worker(), ticket, settings, true, closedBy, reason, &rating)
}
