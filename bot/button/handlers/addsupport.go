package handlers

import (
	"errors"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	cmdregistry "github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AddSupportHandler struct{}

func (h *AddSupportHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "addsupport")
	})
}

func (h *AddSupportHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
		Timeout: time.Second * 30,
	}
}

var addSupportPattern = regexp.MustCompile(`addsupport-(\d)-(\d+)`)

func (h *AddSupportHandler) Execute(ctx *context.ButtonContext) {
	// Permission check
	permLevel, err := ctx.UserPermissionLevel(ctx)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permLevel < permcache.Admin {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
		return
	}

	// Extract data from custom ID
	groups := addSupportPattern.FindStringSubmatch(ctx.InteractionData.CustomId)
	if len(groups) < 3 {
		return
	}

	mentionableTypeRaw, err := strconv.Atoi(groups[1])
	if err != nil {
		return
	}

	mentionableType := context.MentionableType(mentionableTypeRaw)

	id, err := strconv.ParseUint(groups[2], 10, 64)
	if err != nil {
		return
	}

	if mentionableType == context.MentionableTypeUser {
		ctx.ReplyRaw(customisation.Red, "Error", "Users in support teams are now deprecated. Please use roles instead.")
		return

		/* TODO: Remove if Discord does not resolve the performance issues

		// Guild owner doesn't need to be added
		guild, err := ctx.Guild()
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if guild.OwnerId == id {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageOwnerIsAlreadyAdmin)
			return
		}

		if err := dbclient.Client.Permissions.AddSupport(ctx.GuildId(), id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), id, permcache.Support); err != nil {
			ctx.HandleError(err)
			return
		}
		*/
	} else if mentionableType == context.MentionableTypeRole {
		if id == ctx.GuildId() {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageAddSupportEveryone)
			return
		}

		if err := dbclient.Client.RolePermissions.AddSupport(ctx, ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx, ctx.GuildId(), id, permcache.Support); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		ctx.HandleError(fmt.Errorf("invalid mentionable type: %d", mentionableType))
		return
	}

	e := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleAddSupport, i18n.MessageAddSupportSuccess, nil)
	ctx.Edit(command.NewEphemeralEmbedMessageResponse(e))

	updateChannelPermissions(ctx, id, mentionableType)
}

func updateChannelPermissions(ctx cmdregistry.CommandContext, id uint64, mentionableType context.MentionableType) {
	settings, err := ctx.Settings()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if settings.TicketNotificationChannel != nil {
		// Add user / role to thread notification channel
		_ = ctx.Worker().EditChannelPermissions(*settings.TicketNotificationChannel, channel.PermissionOverwrite{
			Id:    id,
			Type:  mentionableType.OverwriteType(),
			Allow: permission.BuildPermissions(permission.ViewChannel, permission.UseApplicationCommands, permission.ReadMessageHistory),
			Deny:  0,
		})
	}

	openTickets, err := dbclient.Client.Tickets.GetGuildOpenTicketsExcludeThreads(ctx, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	panels, err := dbclient.Client.Panel.GetByGuild(ctx, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Update permissions for existing tickets
	for _, ticket := range openTickets {
		if ticket.ChannelId == nil || ticket.IsThread {
			continue
		}

		if ticket.PanelId != nil {
			var panel *database.Panel
			for _, p := range panels {
				if p.PanelId == *ticket.PanelId {
					panel = &p
					break
				}
			}

			if panel == nil {
				continue
			}

			if !panel.WithDefaultTeam {
				continue
			}
		}

		ch, err := ctx.Worker().GetChannel(*ticket.ChannelId)
		if err != nil {
			// Check if the channel has been deleted
			var restError request.RestError
			if errors.As(err, &restError) {
				if restError.StatusCode == 404 {
					if err := dbclient.Client.Tickets.CloseByChannel(ctx, *ticket.ChannelId); err != nil {
						ctx.HandleError(err)
						return
					}

					continue
				} else if restError.StatusCode == 403 {
					break
				}
			}

			continue
		}

		// Apply overwrites to existing channels
		overwrites := append(ch.PermissionOverwrites, channel.PermissionOverwrite{
			Id:    id,
			Type:  mentionableType.OverwriteType(),
			Allow: permission.BuildPermissions(logic.StandardPermissions[:]...),
			Deny:  0,
		})

		data := rest.ModifyChannelData{
			PermissionOverwrites: overwrites,
			Position:             ch.Position,
		}

		if _, err = ctx.Worker().ModifyChannel(*ticket.ChannelId, data); err != nil {
			var restError request.RestError
			if errors.As(err, &restError) {
				if restError.StatusCode == 403 {
					break
				} else if restError.StatusCode == 404 {
					continue
				} else {
					ctx.HandleError(err)
				}
			} else {
				ctx.HandleError(err)
			}

			return
		}
	}
}
