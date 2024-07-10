package handlers

import (
	"fmt"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ViewStaffHandler struct{}

func (h *ViewStaffHandler) Matcher() matcher.Matcher {
	return &matcher.FuncMatcher{
		Func: func(customId string) bool {
			return strings.HasPrefix(customId, "viewstaff_")
		},
	}
}

func (h *ViewStaffHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
		Timeout: time.Second * 5,
	}
}

var viewStaffPattern = regexp.MustCompile(`viewstaff_(\d+)`)

func (h *ViewStaffHandler) Execute(ctx *context.ButtonContext) {
	groups := viewStaffPattern.FindStringSubmatch(ctx.InteractionData.CustomId)
	if len(groups) < 2 {
		return
	}

	page, err := strconv.Atoi(groups[1])
	if err != nil {
		return
	}

	if page < 0 {
		return
	}

	msgEmbed, isBlank := logic.BuildViewStaffMessage(ctx.Context, ctx, page)
	if !isBlank {
		ctx.Edit(command.MessageResponse{
			Embeds: []*embed.Embed{msgEmbed},
			Components: []component.Component{
				component.BuildActionRow(
					component.BuildButton(component.Button{
						CustomId: fmt.Sprintf("viewstaff_%d", page-1),
						Style:    component.ButtonStylePrimary,
						Emoji: &emoji.Emoji{
							Name: "◀️",
						},
						Disabled: page <= 0,
					}),
					component.BuildButton(component.Button{
						CustomId: fmt.Sprintf("viewstaff_%d", page+1),
						Style:    component.ButtonStylePrimary,
						Emoji: &emoji.Emoji{
							Name: "▶️",
						},
						Disabled: false,
					}),
				),
			},
		})
	} else {
		components := ctx.Interaction.Message.Components
		if len(components) == 0 { // Impossible unless whitelabel
			return
		}

		actionRow, ok := components[0].ComponentData.(component.ActionRow)
		if !ok {
			return
		}

		if len(actionRow.Components) < 2 {
			return
		}

		nextButton := actionRow.Components[1].ComponentData.(component.Button)
		if !ok {
			return
		}

		nextButton.Disabled = true
		actionRow.Components[1].ComponentData = nextButton
		components[0].ComponentData = actionRow

		// v hacky
		embeds := make([]*embed.Embed, len(ctx.Interaction.Message.Embeds))
		for i, e := range ctx.Interaction.Message.Embeds {
			embeds[i] = &e
		}

		ctx.Edit(command.MessageResponse{
			Embeds:     embeds,
			Components: components,
		})
	}
}
