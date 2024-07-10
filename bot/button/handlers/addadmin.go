package handlers

import (
	"errors"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
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

type AddAdminHandler struct{}

func (h *AddAdminHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "addadmin")
	})
}

func (h *AddAdminHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
		Timeout: time.Second * 30,
	}
}

var addAdminPattern = regexp.MustCompile(`addadmin-(\d)-(\d+)`)

func (h *AddAdminHandler) Execute(ctx *context.ButtonContext) {
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
	groups := addAdminPattern.FindStringSubmatch(ctx.InteractionData.CustomId)
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

		if err := dbclient.Client.Permissions.AddAdmin(ctx, ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx, ctx.GuildId(), id, permcache.Admin); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.PremiumClient.DeleteCachedTier(ctx, ctx.GuildId()); err != nil {
			ctx.HandleError(err)
			return
		}
	} else if mentionableType == context.MentionableTypeRole {
		if id == ctx.GuildId() {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageAddSupportEveryone)
			return
		}

		if err := dbclient.Client.RolePermissions.AddAdmin(ctx, ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx, ctx.GuildId(), id, permcache.Admin); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		ctx.HandleError(fmt.Errorf("invalid mentionable type: %d", mentionableType))
		return
	}

	e := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleAddAdmin, i18n.MessageAddAdminSuccess, nil)
	ctx.Edit(command.NewEphemeralEmbedMessageResponse(e))

	settings, err := ctx.Settings()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Add user / role to thread notification channel
	if settings.TicketNotificationChannel != nil {
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

	// Update permissions for existing tickets
	for _, ticket := range openTickets {
		if ticket.ChannelId == nil || ticket.IsThread {
			continue
		}

		ch, err := ctx.Worker().GetChannel(*ticket.ChannelId)
		if err != nil {
			// Check if the channel has been deleted
			var restError request.RestError
			if errors.As(err, &restError) && restError.StatusCode == 404 {
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
