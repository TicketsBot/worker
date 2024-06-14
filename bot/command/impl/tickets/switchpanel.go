package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
	"strings"
)

type SwitchPanelCommand struct {
}

func (c SwitchPanelCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "switchpanel",
		Description:     i18n.HelpSwitchPanel,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
		InteractionOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredAutocompleteableArgument("panel", "Ticket panel to switch the ticket to", interaction.OptionTypeInteger, i18n.MessageInvalidUser, c.AutoCompleteHandler), // TODO: Fix invalid message
		),
	}
}

func (c SwitchPanelCommand) GetExecutor() interface{} {
	return c.Execute
}

func (SwitchPanelCommand) Execute(ctx *context.SlashCommandContext, panelId int) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx.Worker().Context, ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 || ticket.ChannelId == nil {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	// Try to move ticket to new category
	panel, err := dbclient.Client.Panel.GetById(panelId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify panel is from same guild
	if panel.PanelId == 0 || panel.GuildId != ctx.GuildId() {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageSwitchPanelInvalidPanel)
		return
	}

	// Update panel assigned to ticket in database
	if err := dbclient.Client.Tickets.SetPanelId(ctx.GuildId(), ticket.Id, panelId); err != nil {
		ctx.HandleError(err)
		return
	}

	// Get ticket claimer
	claimer, err := dbclient.Client.TicketClaims.Get(ticket.GuildId, ticket.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Update welcome message
	if ticket.WelcomeMessageId != nil {
		msg, err := ctx.Worker().GetChannelMessage(ctx.ChannelId(), *ticket.WelcomeMessageId)

		// Error is likely to be due to message being deleted, we want to continue further even if it is
		if err == nil {
			var subject string

			embeds := utils.PtrElems(msg.Embeds) // TODO: Fix types
			if len(embeds) == 0 {
				embeds = make([]*embed.Embed, 1)
				subject = "No subject given"
			} else {
				subject = embeds[0].Title // TODO: Store subjects in database
			}

			embeds[0], err = logic.BuildWelcomeMessageEmbed(ctx, ticket, subject, &panel, nil)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			for i := 1; i < len(embeds); i++ {
				embeds[i].Color = embeds[0].Color
			}

			editData := rest.EditMessageData{
				Content:    msg.Content,
				Embeds:     embeds,
				Flags:      uint(msg.Flags),
				Components: msg.Components,
			}

			if _, err = ctx.Worker().EditMessage(ctx.ChannelId(), *ticket.WelcomeMessageId, editData); err != nil {
				ctx.HandleWarning(err)
			}
		}
	}

	// Get new channel name
	channelName, err := logic.GenerateChannelName(ctx, &panel, ticket.Id, ticket.UserId, utils.NilIfZero(claimer))
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// If the ticket is a thread, we cannot update the permissions (possibly remove a small amount of  members in the
	// future), or the parent channel (user may not have access to it. can you even move threads anyway?)
	if ticket.IsThread {
		settings, err := ctx.Settings()
		if err != nil {
			ctx.HandleError(err)
			return
		}

		data := rest.ModifyChannelData{
			Name: channelName,
		}

		if _, err := ctx.Worker().ModifyChannel(*ticket.ChannelId, data); err != nil {
			ctx.HandleError(err)
			return
		}

		ctx.ReplyRaw(customisation.Green, "Success", "This ticket has been switched to the panel <value here>.\n\nNote: As this is a thread, the permissions could not be bulk updated.")

		// Modify join message
		if ticket.JoinMessageId != nil && settings.TicketNotificationChannel != nil {
			threadStaff, err := logic.GetStaffInThread(ctx.Worker(), ticket, *ticket.ChannelId)
			if err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext()) // Only log
				return
			}

			msg := logic.BuildJoinThreadMessage(ctx.Worker(), ctx.GuildId(), ticket.UserId, ticket.Id, &panel, threadStaff, ctx.PremiumTier())
			if _, err := ctx.Worker().EditMessage(*settings.TicketNotificationChannel, *ticket.JoinMessageId, msg.IntoEditMessageData()); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext()) // Only log
				return
			}
		}

		return
	}

	// Append additional ticket members to overwrites
	members, err := dbclient.Client.TicketMembers.Get(ctx.GuildId(), ticket.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Calculate new channel permissions
	var overwrites []channel.PermissionOverwrite
	if claimer == 0 {
		overwrites, err = logic.CreateOverwrites(ctx, ticket.UserId, &panel, members...)
		if err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		overwrites, err = logic.GenerateClaimedOverwrites(ctx.Worker(), ticket, claimer, members...)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		// GenerateClaimedOverwrites returns nil if the permissions are the same as an unclaimed ticket
		// so if this is the case, we still need to calculate permissions
		if overwrites == nil {
			overwrites, err = logic.CreateOverwrites(ctx, ticket.UserId, &panel, members...)
		}
	}

	// Update channel permissions
	data := rest.ModifyChannelData{
		Name:                 channelName,
		PermissionOverwrites: overwrites,
		ParentId:             panel.TargetCategory,
	}

	if _, err = ctx.Worker().ModifyChannel(*ticket.ChannelId, data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPermanent(customisation.Green, i18n.TitlePanelSwitched, i18n.MessageSwitchPanelSuccess, panel.Title, ctx.UserId())
}

func (SwitchPanelCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice {
	if data.GuildId.Value == 0 {
		return nil
	}

	panels, err := dbclient.Client.Panel.GetByGuild(data.GuildId.Value)
	if err != nil {
		sentry.Error(err) // TODO: Context
		return nil
	}

	if value == "" {
		if len(panels) > 25 {
			return panelsToChoices(panels[:25])
		} else {
			return panelsToChoices(panels)
		}
	} else {
		var filtered []database.Panel
		for _, panel := range panels {
			if strings.Contains(strings.ToLower(panel.Title), strings.ToLower(value)) {
				filtered = append(filtered, panel)
			}

			if len(filtered) == 25 {
				break
			}
		}

		return panelsToChoices(filtered)
	}
}

func panelsToChoices(panels []database.Panel) []interaction.ApplicationCommandOptionChoice {
	choices := make([]interaction.ApplicationCommandOptionChoice, len(panels))
	for i, panel := range panels {
		choices[i] = interaction.ApplicationCommandOptionChoice{
			Name:  panel.Title,
			Value: panel.PanelId,
		}
	}

	return choices
}
