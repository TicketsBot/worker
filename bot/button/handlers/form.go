package handlers

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
	"strings"
)

type FormHandler struct{}

func (h *FormHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "form_")
	})
}

func (h *FormHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed),
		Timeout: constants.TimeoutOpenTicket,
	}
}

func (h *FormHandler) Execute(ctx *context.ModalContext) {
	data := ctx.Interaction.Data
	customId := strings.TrimPrefix(data.CustomId, "form_") // get the custom id that is used in the database

	// Form IDs aren't unique to a panel, so we submit the modal with a custom id of `form_panelcustomid`
	panel, ok, err := dbclient.Client.Panel.GetByCustomId(ctx, ctx.GuildId(), customId)
	if err != nil {
		sentry.Error(err) // TODO: Proper context
		return
	}

	if ok {
		// TODO: Log this
		if panel.GuildId != ctx.GuildId() {
			return
		}

		// blacklist check
		blacklisted, err := ctx.IsBlacklisted(ctx)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if blacklisted {
			ctx.Reply(customisation.Red, i18n.TitleBlacklisted, i18n.MessageBlacklisted)
			return
		}

		inputs, err := dbclient.Client.FormInput.GetAllInputsByCustomId(ctx, ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		formAnswers := make(map[database.FormInput]string)
		for _, actionRow := range data.Components {
			for _, input := range actionRow.Components {
				questionData, ok := inputs[input.CustomId]
				if ok { // If form has changed, we can skip
					formAnswers[questionData] = input.Value
				}
			}
		}

		// Validate user input
		for question, answer := range formAnswers {
			if !question.Required {
				continue
			}

			// Check that users have not just pressed newline or space
			isValid := false
			for _, c := range answer {
				if c != rune(' ') && c != rune('\n') {
					isValid = true
					break
				}
			}

			if !isValid {
				ctx.Reply(customisation.Red, i18n.Error, i18n.MessageFormMissingInput, question.Label)
				return
			}
		}

		ctx.Defer()
		_, _ = logic.OpenTicket(ctx.Context, ctx, &panel, panel.Title, formAnswers)

		return
	}
}
