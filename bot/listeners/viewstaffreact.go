package listeners

import (
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"regexp"
	"strconv"
)

var viewStaffPattern = regexp.MustCompile(`viewstaff_(\d+)`)

func OnViewStaffClick(worker *worker.Context, data interaction.ButtonInteraction, ch chan registry.MessageResponse) {
	// In DMs
	if data.GuildId.Value == 0 {
		return
	}

	groups := viewStaffPattern.FindStringSubmatch(data.Data.CustomId)
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

	errorCtx := errorcontext.WorkerErrorContext{
		Guild:   data.GuildId.Value,
		User:    utils.ButtonInteractionUser(data),
		Channel: data.ChannelId,
	}

	msgEmbed, isBlank := logic.BuildViewStaffMessage(data.GuildId.Value, worker, page, errorCtx)
	if !isBlank {
		ch <- registry.MessageResponse{
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
		}
	} else {
		components := data.Message.Components
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
		embeds := make([]*embed.Embed, len(data.Message.Embeds))
		for i, e := range data.Message.Embeds {
			embeds[i] = &e
		}

		ch <- registry.MessageResponse{
			Embeds:     embeds,
			Components: components,
		}
	}
}
