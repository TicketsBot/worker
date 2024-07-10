package handlers

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
)

type PanelHandler struct{}

func (h *PanelHandler) Matcher() matcher.Matcher {
	return &matcher.DefaultMatcher{}
}

func (h *PanelHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
		Timeout: constants.TimeoutOpenTicket,
	}
}

func (h *PanelHandler) Execute(ctx *context.ButtonContext) {
	panel, ok, err := dbclient.Client.Panel.GetByCustomId(ctx, ctx.GuildId(), ctx.InteractionData.CustomId)
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

		if panel.FormId == nil {
			_, _ = logic.OpenTicket(ctx.Context, ctx, &panel, panel.Title, nil)
		} else {
			form, ok, err := dbclient.Client.Forms.Get(ctx, *panel.FormId)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			if !ok {
				ctx.HandleError(errors.New("Form not found"))
				return
			}

			inputs, err := dbclient.Client.FormInput.GetInputs(ctx, form.Id)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			if len(inputs) == 0 { // Don't open a blank form
				_, _ = logic.OpenTicket(ctx.Context, ctx, &panel, panel.Title, nil)
			} else {
				modal := buildForm(panel, form, inputs)
				ctx.Modal(modal)
			}
		}

		return
	}
}

func buildForm(panel database.Panel, form database.Form, inputs []database.FormInput) button.ResponseModal {
	components := make([]component.Component, len(inputs))
	for i, input := range inputs {
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
		}))
	}

	return button.ResponseModal{
		Data: interaction.ModalResponseData{
			CustomId:   fmt.Sprintf("form_%s", panel.CustomId),
			Title:      form.Title,
			Components: components,
		},
	}
}
