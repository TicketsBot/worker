package handlers

import (
	"fmt"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/logic"
)

type CloseWithReasonSubmitHandler struct{}

func (h *CloseWithReasonSubmitHandler) Matcher() matcher.Matcher {
	return matcher.NewSimpleMatcher("close_with_reason_submit")
}

func (h *CloseWithReasonSubmitHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed),
		Timeout: constants.TimeoutCloseTicket,
	}
}

func (h *CloseWithReasonSubmitHandler) Execute(ctx *context.ModalContext) {
	data := ctx.Interaction.Data

	// Get the reason
	if len(data.Components) == 0 { // No action rows
		ctx.HandleError(fmt.Errorf("No action rows found in modal components"))
		return
	}

	actionRow := data.Components[0]
	if len(actionRow.Components) == 0 { // Text input missing
		ctx.HandleError(fmt.Errorf("Modal missing text input"))
		return
	}

	textInput := actionRow.Components[0]
	if textInput.CustomId != "reason" {
		ctx.HandleError(fmt.Errorf("Text input custom ID mismatch"))
		return
	}

	// This must be malicious
	if len(textInput.Value) > 1024 {
		ctx.HandleError(fmt.Errorf("Reason is too long"))
		return
	}

	ctx.Ack()
	logic.CloseTicket(ctx.Context, ctx, &textInput.Value, false)
}
