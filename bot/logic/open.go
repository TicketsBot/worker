package logic

import (
	"context"
	"errors"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/permissionwrapper"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"golang.org/x/sync/errgroup"
	"sync"
)

func OpenTicket(ctx registry.CommandContext, panel *database.Panel, subject string) {
	// If we're using a panel, then we need to create the ticket in the specified category
	var category uint64
	if panel != nil && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		var err error
		category, err = dbclient.Client.ChannelCategory.Get(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return
		}
	}

	// TODO: Re-add permission check
	/*requiredPerms := []permission.Permission{
		permission.ManageChannels,
		permission.ManageRoles,
		permission.ViewChannel,
		permission.SendMessages,
		permission.ReadMessageHistory,
	}

	if !permission.HasPermissions(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), requiredPerms...) {
		ctx.Reply(utils.Red, "Error", "I am missing the required permissions. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
		if ctx.ShouldReact {
			ctx.ReactWithCross()
		}
		return
	}*/

	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		_, err := ctx.Worker().GetChannel(category)
		if err != nil {
			useCategory = false

			if restError, ok := err.(request.RestError); ok && restError.StatusCode == 404 {
				if panel == nil {
					if err := dbclient.Client.ChannelCategory.Delete(ctx.GuildId()); err != nil {
						ctx.HandleError(err)
					}
				} // TODO: Else, set panel category to 0
			}
		} else {
			// TODO: Re-add permission check
			/*if !permission.HasPermissionsChannel(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), category, requiredPerms...) {
				ctx.Reply(utils.Red, "Error", "I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
				if ctx.ShouldReact {
					ctx.ReactWithCross()
				}
				return
			}*/
		}
	}

	var targetChannel uint64

	// Make sure ticket count is within ticket limit
	violatesTicketLimit, limit := getTicketLimit(ctx)
	if violatesTicketLimit {
		// initialise target channel
		if targetChannel == 0 {
			var err error
			targetChannel, err = getErrorTargetChannel(ctx, panel)
			if err != nil {
				ctx.HandleError(err)
				return
			}
		}

		// Notify the user
		if targetChannel != 0 {
			ticketsPluralised := "ticket"
			if limit > 1 {
				ticketsPluralised += "s"
			}

			ctx.Reply(utils.Red, "Error", i18n.MessageTicketLimitReached, limit, ticketsPluralised)
		}

		return
	}

	// Generate subject
	if panel != nil && panel.Title != "" { // If we're using a panel, use the panel title as the subject
		subject = panel.Title
	} else { // Else, take command args as the subject
		if subject == "" {
			subject = "No subject given"
		}

		if len(subject) > 256 {
			subject = subject[0:255]
		}
	}

	// Make sure there's not > 50 channels in a category
	if useCategory {
		channels, _ := ctx.Worker().GetGuildChannels(ctx.GuildId())

		channelCount := 0
		for _, channel := range channels {
			if channel.ParentId.Value == category {
				channelCount++
			}
		}

		if channelCount >= 50 {
			ctx.Reply(utils.Red, "Error", i18n.MessageTooManyTickets)
			return
		}
	}

	// Create channel
	id, err := dbclient.Client.Tickets.Create(ctx.GuildId(), ctx.UserId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	overwrites := CreateOverwrites(ctx.Worker(), ctx.GuildId(), ctx.UserId(), ctx.Worker().BotId, panel)

	// Create ticket name
	var name string

	namingScheme, err := dbclient.Client.NamingScheme.Get(ctx.GuildId())
	if err != nil {
		namingScheme = database.Id
		ctx.HandleError(err)
	}

	if namingScheme == database.Username {
		user, err := ctx.User()
		if err != nil {
			ctx.HandleError(err)
			return
		}

		name = fmt.Sprintf("ticket-%s", user.Username)
	} else {
		name = fmt.Sprintf("ticket-%d", id)
	}

	data := rest.CreateChannelData{
		Name:                 name,
		Type:                 channel.ChannelTypeGuildText,
		Topic:                subject,
		PermissionOverwrites: overwrites,
		ParentId:             category,
	}
	if useCategory {
		data.ParentId = category
	}

	channel, err := ctx.Worker().CreateGuildChannel(ctx.GuildId(), data)
	if err != nil { // Bot likely doesn't have permission
		ctx.HandleError(err)

		// To prevent tickets getting in a glitched state, we should mark it as closed (or delete it completely?)
		if err := dbclient.Client.Tickets.Close(id, ctx.GuildId()); err != nil {
			ctx.HandleError(err)
		}

		return
	}

	ctx.Accept()

	welcomeMessageId, err := utils.SendWelcomeMessage(ctx.Worker(), ctx.GuildId(), channel.Id, ctx.UserId(), ctx.PremiumTier() > premium.None, subject, panel, id)
	if err != nil {
		ctx.HandleError(err)
	}

	var panelId *int
	if panel != nil {
		panelId = &panel.PanelId
	}

	// UpdateUser channel in DB
	if err := dbclient.Client.Tickets.SetTicketProperties(ctx.GuildId(), id, channel.Id, welcomeMessageId, panelId); err != nil {
		ctx.HandleError(err)
	}

	// mentions
	{
		var content string

		if panel != nil {
			// roles
			roles, err := dbclient.Client.PanelRoleMentions.GetRoles(panel.PanelId)
			if err != nil {
				ctx.HandleError(err)
			} else {
				for _, roleId := range roles {
					if roleId == ctx.GuildId() {
						content += "@everyone"
					} else {
						content += fmt.Sprintf("<@&%d>", roleId)
					}
				}
			}

			// user
			shouldMentionUser, err := dbclient.Client.PanelUserMention.ShouldMentionUser(panel.PanelId)
			if err != nil {
				ctx.HandleError(err)
			} else {
				if shouldMentionUser {
					content += fmt.Sprintf("<@%d>", ctx.UserId())
				}
			}
		}

		if content != "" {
			if len(content) > 2000 {
				content = content[:2000]
			}

			pingMessage, err := ctx.Worker().CreateMessageComplex(channel.Id, rest.CreateMessageData{
				Content: content,
				AllowedMentions: message.AllowedMention{
					Parse: []message.AllowedMentionType{
						message.EVERYONE,
						message.USERS,
						message.ROLES,
					},
				},
			})

			if err != nil {
				ctx.HandleError(err)
			} else {
				// error is likely to be a permission error
				_ = ctx.Worker().DeleteMessage(channel.Id, pingMessage.Id)
			}
		}
	}

	// Let the user know the ticket has been opened
	if panel == nil {
		ctx.Reply(utils.Green, "Ticket", i18n.MessageTicketOpened, channel.Mention())
	}
	/*else {
		dmOnOpen, err := dbclient.Client.DmOnOpen.Get(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
		}

		if dmOnOpen && dmChannel.Id != 0 {
			ctx.Reply(utils.Green, "Ticket", i18n.MessageTicketOpened, channel.Mention())
		}
	}*/

	go statsd.Client.IncrementKey(statsd.KeyTickets)

	if ctx.PremiumTier() > premium.None {
		go createWebhook(ctx.Worker(), id, ctx.GuildId(), channel.Id)
	}

	// update cache
	go func() {
		// retrieve member
		// GetGuildMember will cache if not already cached
		if _, err := ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId()); err != nil {
			ctx.HandleError(err)
		}

		// cache user
		if _, err := ctx.Worker().GetUser(ctx.UserId()); err != nil {
			ctx.HandleError(err)
		}
	}()
}

// has hit ticket limit, ticket limit
func getTicketLimit(ctx registry.CommandContext) (bool, int) {
	isStaff, err := ctx.UserPermissionLevel()
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return true, 1 // TODO: Stop flow
	}

	if isStaff >= permcache.Support {
		return false, 50
	}

	var openedTickets []database.Ticket
	var ticketLimit uint8

	group, _ := errgroup.WithContext(context.Background())

	// get ticket limit
	group.Go(func() (err error) {
		ticketLimit, err = dbclient.Client.TicketLimit.Get(ctx.GuildId())
		return
	})

	group.Go(func() (err error) {
		openedTickets, err = dbclient.Client.Tickets.GetOpenByUser(ctx.GuildId(), ctx.UserId())
		return
	})

	if err := group.Wait(); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return true, 1
	}

	return len(openedTickets) >= int(ticketLimit), int(ticketLimit)
}

func createWebhook(worker *worker.Context, ticketId int, guildId, channelId uint64) {
	// TODO: Re-add permission check
	//if permission.HasPermissionsChannel(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), channelId, permission.ManageWebhooks) { // Do we actually need this?

	var data rest.WebhookData

	self, err := worker.Self()
	if err == nil {
		data = rest.WebhookData{
			Username: self.Username,
			Avatar:   self.AvatarUrl(256),
		}
	} else {
		data = rest.WebhookData{
			Username: "Tickets",
		}
	}

	webhook, err := worker.CreateWebhook(channelId, data)
	if err != nil {
		sentry.Error(err)
		return
	}

	dbWebhook := database.Webhook{
		Id:    webhook.Id,
		Token: webhook.Token,
	}

	if err := dbclient.Client.Webhooks.Create(guildId, ticketId, dbWebhook); err != nil {
		sentry.Error(err)
	}
}

var allowedPermissions = []permission.Permission{permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks}

func CreateOverwrites(worker *worker.Context, guildId, userId, selfId uint64, panel *database.Panel) (overwrites []channel.PermissionOverwrite) {
	errorContext := errorcontext.WorkerErrorContext{
		Guild: guildId,
		User:  userId,
	}

	// Apply permission overwrites
	overwrites = append(overwrites, channel.PermissionOverwrite{ // @everyone
		Id:    guildId,
		Type:  channel.PermissionTypeRole,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel),
	})

	// Create list of members & roles who should be added to the ticket
	allowedUsers := make([]uint64, 0)
	allowedRoles := make([]uint64, 0)

	// Should we add the default team
	if panel == nil || panel.WithDefaultTeam {
		// Get support reps & admins
		supportUsers, err := dbclient.Client.Permissions.GetSupport(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}

		allowedUsers = append(allowedUsers, supportUsers...)

		// Get support roles & admin roles
		supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}

		allowedRoles = append(allowedUsers, supportRoles...)
	}

	// Add other support teams
	if panel != nil {
		teams, err := dbclient.Client.PanelTeams.GetTeams(panel.PanelId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
		} else {
			group, _ := errgroup.WithContext(context.Background())
			mu := sync.Mutex{}

			for _, team := range teams {
				team := team

				// TODO: Joins
				group.Go(func() error {
					members, err := dbclient.Client.SupportTeamMembers.Get(team.Id)
					if err != nil {
						return err
					}

					roles, err := dbclient.Client.SupportTeamRoles.Get(team.Id)
					if err != nil {
						return err
					}

					mu.Lock()
					defer mu.Unlock()
					allowedUsers = append(allowedUsers, members...)
					allowedRoles = append(allowedRoles, roles...)

					return nil
				})
			}

			if err := group.Wait(); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		}
	}

	// Add the sender & self
	allowedUsers = append(allowedUsers, userId, selfId)

	for _, member := range allowedUsers {
		allow := allowedPermissions

		// Give ourselves permissions to create webhooks
		if member == selfId {
			if permissionwrapper.HasPermissions(worker, guildId, selfId, permission.ManageWebhooks) {
				allow = append(allowedPermissions, permission.ManageWebhooks)
			}
		}

		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    member,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(allow...),
			Deny:  0,
		})
	}

	for _, role := range allowedRoles {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    role,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(allowedPermissions...),
			Deny:  0,
		})
	}

	return overwrites
}

// target channel for messaging the user
// either DMs or the channel where the command was run
func getErrorTargetChannel(ctx registry.CommandContext, panel *database.Panel) (uint64, error) {
	if panel == nil {
		return ctx.ChannelId(), nil
	} else {
		dmChannel, ok := getDmChannel(ctx, ctx.UserId())
		if !ok {
			return 0, errors.New("failed to create dm channel")
		}

		return dmChannel, nil
	}
}