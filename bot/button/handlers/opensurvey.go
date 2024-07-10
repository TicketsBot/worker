package handlers

import (
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type OpenSurveyHandler struct{}

func (h *OpenSurveyHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "open-exit-survey-")
	})
}

func (h *OpenSurveyHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.DMsAllowed),
		Timeout: time.Second * 3,
	}
}

var openSurveyPattern = regexp.MustCompile(`open-exit-survey-(\d+)-(\d+)`)

func (h *OpenSurveyHandler) Execute(ctx *context.ButtonContext) {
	groups := openSurveyPattern.FindStringSubmatch(ctx.InteractionData.CustomId)
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

	premiumTier, err := utils.PremiumClient.GetTierByGuildId(ctx, guildId, true, ctx.Worker().Token, ctx.Worker().RateLimiter)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if premiumTier == premium.None {
		ctx.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	// Get ticket
	ticket, err := dbclient.Client.Tickets.Get(ctx, ticketId, guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.UserId != ctx.InteractionUser().Id || ticket.GuildId != guildId || ticket.Id != ticketId {
		return
	}

	feedbackEnabled, err := dbclient.Client.FeedbackEnabled.Get(ctx, guildId)
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

	panel, err := dbclient.Client.Panel.GetById(ctx, *ticket.PanelId)
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

	form, ok, err := dbclient.Client.Forms.Get(ctx, *panel.ExitSurveyFormId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !ok {
		ctx.ReplyRaw(customisation.Red, "Error", "The survey is no longer available for this ticket.") // TODO: i18n
		return
	}

	formInputs, err := dbclient.Client.FormInput.GetInputs(ctx, form.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	components := make([]component.Component, len(formInputs))
	for i, input := range formInputs {
		var minLength, maxLength *uint32
		if input.MinLength != nil && *input.MinLength > 0 {
			minLength = utils.Ptr(uint32(*input.MinLength))
		}

		if input.MaxLength != nil {
			maxLength = utils.Ptr(uint32(*input.MaxLength))
		}

		components[i] = component.BuildActionRow(component.BuildInputText(component.InputText{
			Style:       component.TextStyleTypes(input.Style),
			CustomId:    input.CustomId,
			Label:       input.Label,
			Placeholder: input.Placeholder,
			MinLength:   minLength,
			MaxLength:   maxLength,
			Required:    utils.Ptr(input.Required),
			Value:       nil,
		}))
	}

	ctx.Modal(button.ResponseModal{
		Data: interaction.ModalResponseData{
			CustomId:   fmt.Sprintf("exit-survey-%d-%d", guildId, ticketId),
			Title:      form.Title,
			Components: components,
		},
	})
}
