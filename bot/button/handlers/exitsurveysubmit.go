package handlers

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ExitSurveySubmitHandler struct{}

func (h *ExitSurveySubmitHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "exit-survey-")
	})
}

func (h *ExitSurveySubmitHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.DMsAllowed),
		Timeout: time.Second * 8,
	}
}

var exitSurveyPattern = regexp.MustCompile(`exit-survey-(\d+)-(\d+)`)

func (h *ExitSurveySubmitHandler) Execute(cmd *cmdcontext.ModalContext) {
	groups := exitSurveyPattern.FindStringSubmatch(cmd.Interaction.Data.CustomId)
	if len(groups) != 3 {
		return
	}

	ctx, cancel := context.WithTimeout(cmd.Context, time.Second*10)
	defer cancel()

	// Error may occur if guild ID in custom ID > max u64 size
	guildId, err := strconv.ParseUint(groups[1], 10, 64)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	ticketId, err := strconv.Atoi(groups[2])
	if err != nil {
		cmd.HandleError(err)
		return
	}

	premiumTier, err := utils.PremiumClient.GetTierByGuildId(ctx, guildId, true, cmd.Worker().Token, cmd.Worker().RateLimiter)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if premiumTier == premium.None {
		cmd.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	// Get ticket
	ticket, err := dbclient.Client.Tickets.Get(ctx, ticketId, guildId)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if ticket.UserId != cmd.InteractionUser().Id || ticket.GuildId != guildId || ticket.Id != ticketId {
		return
	}

	feedbackEnabled, err := dbclient.Client.FeedbackEnabled.Get(ctx, guildId)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if !feedbackEnabled {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageFeedbackDisabled)
		return
	}

	if ticket.PanelId == nil {
		cmd.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	panel, err := dbclient.Client.Panel.GetById(ctx, *ticket.PanelId)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if panel.GuildId != guildId || panel.PanelId != *ticket.PanelId {
		cmd.HandleError(fmt.Errorf("panel not found"))
		return
	}

	if panel.ExitSurveyFormId == nil {
		cmd.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	form, ok, err := dbclient.Client.Forms.Get(ctx, *panel.ExitSurveyFormId)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if !ok {
		cmd.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	formInputs, err := dbclient.Client.FormInput.GetInputs(ctx, form.Id)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	responses := make(map[int]string)
	for _, input := range formInputs {
		value, ok := cmd.GetInput(input.CustomId)
		if ok {
			responses[input.Id] = value
		}
	}

	if err := dbclient.Client.ExitSurveyResponses.AddResponses(ctx, guildId, ticketId, *panel.ExitSurveyFormId, responses); err != nil {
		cmd.HandleError(err)
		return
	}

	cmd.EditWithRaw(customisation.Green, "Success", "Thank you for your feedback!") // TODO: i18n

	if err := addViewFeedbackButton(ctx, cmd, ticket); err != nil {
		cmd.HandleError(err)
		return
	}
}

func addViewFeedbackButton(ctx context.Context, cmd *cmdcontext.ModalContext, ticket database.Ticket) error {
	// Get archive message
	settings, err := cmd.Settings()
	if err != nil {
		return err
	}

	closeMetadata, ok, err := dbclient.Client.CloseReason.Get(ctx, ticket.GuildId, ticket.Id)
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

	rating, ok, err := dbclient.Client.ServiceRatings.Get(ctx, ticket.GuildId, ticket.Id)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("exit survey was completed, but no rating was found (%d:%d)", ticket.GuildId, ticket.Id)
	}

	return logic.EditGuildArchiveMessageIfExists(ctx, cmd.Worker(), ticket, settings, true, closedBy, reason, &rating)
}
