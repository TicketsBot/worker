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
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"time"
)

func OpenTicket(ctx context.Context, cmd registry.InteractionContext, panel *database.Panel, subject string, formData map[database.FormInput]string) (database.Ticket, error) {
	rootSpan := sentry.StartSpan(ctx, "Ticket open")
	rootSpan.SetTag("guild", strconv.FormatUint(cmd.GuildId(), 10))
	defer rootSpan.Finish()

	span := sentry.StartSpan(rootSpan.Context(), "Check ticket limit")

	lockCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	mu, err := redis.TakeTicketOpenLock(lockCtx, cmd.GuildId())
	if err != nil {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}

	unlocked := false
	defer func() {
		if !unlocked {
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()

			mu.UnlockContext(ctx)
		}
	}()

	// Make sure ticket count is within ticket limit
	// Check ticket limit before ratelimit token to prevent 1 person from stopping everyone opening tickets
	violatesTicketLimit, limit := getTicketLimit(ctx, cmd)
	if violatesTicketLimit {
		// Notify the user
		ticketsPluralised := "ticket"
		if limit > 1 {
			ticketsPluralised += "s"
		}

		// TODO: Use translation of tickets
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageTicketLimitReached, limit, ticketsPluralised)
		return database.Ticket{}, fmt.Errorf("ticket limit reached")
	}

	span.Finish()

	span = sentry.StartSpan(rootSpan.Context(), "Ticket ratelimit")

	ok, err := redis.TakeTicketRateLimitToken(redis.Client, cmd.GuildId())
	if err != nil {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}

	span.Finish()

	if !ok {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageOpenRatelimited)
		return database.Ticket{}, nil
	}

	// Ensure that the panel isn't disabled
	span = sentry.StartSpan(rootSpan.Context(), "Check if panel is disabled")
	if panel != nil && panel.ForceDisabled {
		// Build premium command mention
		var premiumCommand string
		commands, err := command.LoadCommandIds(cmd.Worker(), cmd.Worker().BotId)
		if err != nil {
			sentry.Error(err)
			return database.Ticket{}, err
		}

		if id, ok := commands["premium"]; ok {
			premiumCommand = fmt.Sprintf("</premium:%d>", id)
		} else {
			premiumCommand = "`/premium`"
		}

		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageOpenPanelForceDisabled, premiumCommand)
		return database.Ticket{}, nil
	}

	span.Finish()

	if panel != nil && panel.Disabled {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageOpenPanelDisabled)
		return database.Ticket{}, nil
	}

	if panel != nil {
		member, err := cmd.Member()
		if err != nil {
			cmd.HandleError(err)
			return database.Ticket{}, err
		}

		matchedRole, action, err := dbclient.Client.PanelAccessControlRules.GetFirstMatched(
			ctx,
			panel.PanelId,
			append(member.Roles, cmd.GuildId()),
		)

		if err != nil {
			cmd.HandleError(err)
			return database.Ticket{}, err
		}

		if action == database.AccessControlActionDeny {
			if err := sendAccessControlDeniedMessage(ctx, cmd, panel.PanelId, matchedRole); err != nil {
				cmd.HandleError(err)
				return database.Ticket{}, err
			}

			return database.Ticket{}, nil
		} else if action != database.AccessControlActionAllow {
			cmd.HandleError(fmt.Errorf("invalid access control action %s", action))
			return database.Ticket{}, err
		}
	}

	span = sentry.StartSpan(rootSpan.Context(), "Load settings")
	settings, err := cmd.Settings()
	if err != nil {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}
	span.Finish()

	isThread := settings.UseThreads

	// Check if the parent channel is an announcement channel
	span = sentry.StartSpan(rootSpan.Context(), "Check if parent channel is announcement channel")
	if isThread {
		panelChannel, err := cmd.Channel()
		if err != nil {
			cmd.HandleError(err)
			return database.Ticket{}, err
		}

		if panelChannel.Type != channel.ChannelTypeGuildText {
			cmd.Reply(customisation.Red, i18n.Error, i18n.MessageOpenThreadAnnouncementChannel)
			return database.Ticket{}, nil
		}
	}
	span.Finish()

	// If we're using a panel, then we need to create the ticket in the specified category
	span = sentry.StartSpan(rootSpan.Context(), "Get category")
	var category uint64
	if panel != nil && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		var err error
		category, err = dbclient.Client.ChannelCategory.Get(ctx, cmd.GuildId())
		if err != nil {
			cmd.HandleError(err)
			return database.Ticket{}, err
		}
	}
	span.Finish()

	useCategory := category != 0 && !isThread
	if useCategory {
		span := sentry.StartSpan(rootSpan.Context(), "Check if category exists")
		// Check if the category still exists
		_, err := cmd.Worker().GetChannel(category)
		if err != nil {
			useCategory = false

			if restError, ok := err.(request.RestError); ok && restError.StatusCode == 404 {
				if panel == nil {
					if err := dbclient.Client.ChannelCategory.Delete(ctx, cmd.GuildId()); err != nil {
						cmd.HandleError(err)
					}
				} // TODO: Else, set panel category to 0
			}
		}
		span.Finish()
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

	// Channel count checks
	if !isThread {
		span := sentry.StartSpan(rootSpan.Context(), "Check < 500 channels")
		channels, _ := cmd.Worker().GetGuildChannels(cmd.GuildId())

		// 500 guild limit check
		if !isThread && countRealChannels(channels, 0) >= 500 {
			cmd.Reply(customisation.Red, i18n.Error, i18n.MessageGuildChannelLimitReached)
			return database.Ticket{}, fmt.Errorf("channel limit reached")
		}

		span.Finish()

		// Make sure there's not > 50 channels in a category
		if useCategory {
			span := sentry.StartSpan(rootSpan.Context(), "Check < 50 channels in category")
			categoryChildrenCount := countRealChannels(channels, category)

			if categoryChildrenCount >= 50 {
				// Try to use the overflow category if there is one
				if settings.OverflowEnabled {
					// If overflow is enabled, and the category id is nil, then use the root of the server
					if settings.OverflowCategoryId == nil {
						useCategory = false
					} else {
						category = *settings.OverflowCategoryId

						// Verify that the overflow category still exists
						span := sentry.StartSpan(span.Context(), "Check if overflow category exists")
						if _, err := cmd.Worker().GetChannel(category); err != nil {
							var restError request.RestError
							if errors.As(err, &restError) && restError.StatusCode == 404 {
								if err := dbclient.Client.Settings.SetOverflow(ctx, cmd.GuildId(), false, nil); err != nil {
									cmd.HandleError(err)
									return database.Ticket{}, err
								}
							}

							cmd.Reply(customisation.Red, i18n.Error, i18n.MessageTooManyTickets)
							return database.Ticket{}, err
						}

						// Check that the overflow category still has space
						overflowCategoryChildrenCount := countRealChannels(channels, *settings.OverflowCategoryId)

						if overflowCategoryChildrenCount >= 50 {
							cmd.Reply(customisation.Red, i18n.Error, i18n.MessageTooManyTickets)
							return database.Ticket{}, fmt.Errorf("overflow category full")
						}

						span.Finish()
					}
				} else {
					cmd.Reply(customisation.Red, i18n.Error, i18n.MessageTooManyTickets)
					return database.Ticket{}, fmt.Errorf("category ticket limit reached")
				}
			}

			span.Finish()
		}
	}

	var panelId *int
	if panel != nil {
		panelId = &panel.PanelId
	}

	// Create channel
	span = sentry.StartSpan(rootSpan.Context(), "Create ticket in database")
	ticketId, err := dbclient.Client.Tickets.Create(ctx, cmd.GuildId(), cmd.UserId(), isThread, panelId)
	if err != nil {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}
	span.Finish()

	unlocked = true
	if _, err := mu.UnlockContext(ctx); err != nil && !errors.Is(err, redis.ErrLockExpired) {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}

	span = sentry.StartSpan(rootSpan.Context(), "Generate channel name")
	name, err := GenerateChannelName(ctx, cmd, panel, ticketId, cmd.UserId(), nil)
	if err != nil {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}
	span.Finish()

	var ch channel.Channel
	var joinMessageId *uint64
	if isThread {
		span = sentry.StartSpan(rootSpan.Context(), "Create thread")
		ch, err = cmd.Worker().CreatePrivateThread(cmd.ChannelId(), name, uint16(settings.ThreadArchiveDuration), false)
		if err != nil {
			cmd.HandleError(err)

			// To prevent tickets getting in a glitched state, we should mark it as closed (or delete it completely?)
			if err := dbclient.Client.Tickets.Close(ctx, ticketId, cmd.GuildId()); err != nil {
				cmd.HandleError(err)
			}

			return database.Ticket{}, err
		}
		span.Finish()

		// Join ticket
		span = sentry.StartSpan(rootSpan.Context(), "Add user to thread")
		if err := cmd.Worker().AddThreadMember(ch.Id, cmd.UserId()); err != nil {
			cmd.HandleError(err)
		}
		span.Finish()

		if settings.TicketNotificationChannel != nil {
			span := sentry.StartSpan(rootSpan.Context(), "Send message to ticket notification channel")

			buildSpan := sentry.StartSpan(span.Context(), "Build ticket notification message")
			data := BuildJoinThreadMessage(ctx, cmd.Worker(), cmd.GuildId(), cmd.UserId(), ticketId, panel, nil, cmd.PremiumTier())
			buildSpan.Finish()

			// TODO: Check if channel exists
			if msg, err := cmd.Worker().CreateMessageComplex(*settings.TicketNotificationChannel, data.IntoCreateMessageData()); err == nil {
				joinMessageId = &msg.Id
			} else {
				cmd.HandleError(err)
			}
			span.Finish()
		}
	} else {
		span = sentry.StartSpan(rootSpan.Context(), "Build permission overwrites")
		overwrites, err := CreateOverwrites(ctx, cmd, cmd.UserId(), panel)
		if err != nil {
			cmd.HandleError(err)
			return database.Ticket{}, err
		}
		span.Finish()

		data := rest.CreateChannelData{
			Name:                 name,
			Type:                 channel.ChannelTypeGuildText,
			Topic:                subject,
			PermissionOverwrites: overwrites,
		}

		if useCategory {
			data.ParentId = category
		}

		span = sentry.StartSpan(rootSpan.Context(), "Create channel")
		tmp, err := cmd.Worker().CreateGuildChannel(cmd.GuildId(), data)
		if err != nil { // Bot likely doesn't have permission
			cmd.HandleError(err)

			// To prevent tickets getting in a glitched state, we should mark it as closed (or delete it completely?)
			if err := dbclient.Client.Tickets.Close(ctx, ticketId, cmd.GuildId()); err != nil {
				cmd.HandleError(err)
			}

			return database.Ticket{}, err
		}
		span.Finish()

		// TODO: Remove
		if tmp.Id == 0 {
			cmd.HandleError(fmt.Errorf("channel id is 0"))
			return database.Ticket{}, fmt.Errorf("channel id is 0")
		}

		ch = tmp
	}

	if err := dbclient.Client.Tickets.SetChannelId(ctx, cmd.GuildId(), ticketId, ch.Id); err != nil {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}

	prometheus.LogTicketCreated(cmd.GuildId())

	// Parallelise as much as possible
	group, _ := errgroup.WithContext(ctx)

	// Let the user know the ticket has been opened
	group.Go(func() error {
		span := sentry.StartSpan(rootSpan.Context(), "Reply to interaction")
		cmd.Reply(customisation.Green, i18n.Ticket, i18n.MessageTicketOpened, ch.Mention())
		span.Finish()
		return nil
	})

	// WelcomeMessageId is modified in the welcome message goroutine
	ticket := database.Ticket{
		Id:               ticketId,
		GuildId:          cmd.GuildId(),
		ChannelId:        &ch.Id,
		UserId:           cmd.UserId(),
		Open:             true,
		OpenTime:         time.Now(), // will be a bit off, but not used
		WelcomeMessageId: nil,
		PanelId:          panelId,
		IsThread:         isThread,
		JoinMessageId:    joinMessageId,
	}

	// Welcome message
	group.Go(func() error {
		span = sentry.StartSpan(rootSpan.Context(), "Fetch custom integration placeholders")

		externalPlaceholderCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		additionalPlaceholders, err := fetchCustomIntegrationPlaceholders(externalPlaceholderCtx, ticket, formAnswersToMap(formData))
		if err != nil {
			// TODO: Log for integration author and server owner on the dashboard, rather than spitting out a message.
			// A failing integration should not block the ticket creation process.
			cmd.HandleError(err)
		}
		span.Finish()

		span = sentry.StartSpan(rootSpan.Context(), "Send welcome message")
		welcomeMessageId, err := SendWelcomeMessage(ctx, cmd, ticket, subject, panel, formData, additionalPlaceholders)
		span.Finish()
		if err != nil {
			return err
		}

		// Update message IDs in DB
		span = sentry.StartSpan(rootSpan.Context(), "Update ticket properties in database")
		defer span.Finish()
		if err := dbclient.Client.Tickets.SetMessageIds(ctx, cmd.GuildId(), ticketId, welcomeMessageId, joinMessageId); err != nil {
			return err
		}

		return nil
	})

	// Send mentions
	group.Go(func() error {
		span := sentry.StartSpan(rootSpan.Context(), "Load guild metadata from database")
		metadata, err := dbclient.Client.GuildMetadata.Get(ctx, cmd.GuildId())
		span.Finish()
		if err != nil {
			return err
		}

		// mentions
		var content string

		// Append on-call role pings
		if isThread {
			if panel == nil {
				if metadata.OnCallRole != nil {
					content += fmt.Sprintf("<@&%d>", *metadata.OnCallRole)
				}
			} else {
				if panel.WithDefaultTeam && metadata.OnCallRole != nil {
					content += fmt.Sprintf("<@&%d>", *metadata.OnCallRole)
				}

				span := sentry.StartSpan(rootSpan.Context(), "Get teams from database")
				teams, err := dbclient.Client.PanelTeams.GetTeams(ctx, panel.PanelId)
				span.Finish()
				if err != nil {
					return err
				} else {
					for _, team := range teams {
						if team.OnCallRole != nil {
							content += fmt.Sprintf("<@&%d>", *team.OnCallRole)
						}
					}
				}
			}
		}

		if panel != nil {
			// roles
			span := sentry.StartSpan(rootSpan.Context(), "Get panel role mentions from database")
			roles, err := dbclient.Client.PanelRoleMentions.GetRoles(ctx, panel.PanelId)
			span.Finish()
			if err != nil {
				return err
			} else {
				for _, roleId := range roles {
					if roleId == cmd.GuildId() {
						content += "@everyone"
					} else {
						content += fmt.Sprintf("<@&%d>", roleId)
					}
				}
			}

			// user
			span = sentry.StartSpan(rootSpan.Context(), "Get panel user mention setting from database")
			shouldMentionUser, err := dbclient.Client.PanelUserMention.ShouldMentionUser(ctx, panel.PanelId)
			span.Finish()
			if err != nil {
				return err
			} else {
				if shouldMentionUser {
					content += fmt.Sprintf("<@%d>", cmd.UserId())
				}
			}
		}

		if content != "" {
			if len(content) > 2000 {
				content = content[:2000]
			}

			span := sentry.StartSpan(rootSpan.Context(), "Send ping message")
			pingMessage, err := cmd.Worker().CreateMessageComplex(ch.Id, rest.CreateMessageData{
				Content: content,
				AllowedMentions: message.AllowedMention{
					Parse: []message.AllowedMentionType{
						message.EVERYONE,
						message.USERS,
						message.ROLES,
					},
				},
			})
			span.Finish()

			if err != nil {
				return err
			}

			// error is likely to be a permission error
			span = sentry.StartSpan(span.Context(), "Delete ping message")
			_ = cmd.Worker().DeleteMessage(ch.Id, pingMessage.Id)
			span.Finish()
		}

		return nil
	})

	// Create webhook
	// TODO: Create webhook on use, rather than on ticket creation.
	// TODO: Webhooks for threads should be created on the parent channel.
	if cmd.PremiumTier() > premium.None && !ticket.IsThread {
		group.Go(func() error {
			return createWebhook(rootSpan.Context(), cmd, ticketId, cmd.GuildId(), ch.Id)
		})
	}

	if err := group.Wait(); err != nil {
		cmd.HandleError(err)
		return database.Ticket{}, err
	}

	span = sentry.StartSpan(rootSpan.Context(), "Increment statsd counters")
	statsd.Client.IncrementKey(statsd.KeyTickets)
	if panel == nil {
		statsd.Client.IncrementKey(statsd.KeyOpenCommand)
	}
	span.Finish()

	return ticket, nil
}

// has hit ticket limit, ticket limit
func getTicketLimit(ctx context.Context, cmd registry.CommandContext) (bool, int) {
	isStaff, err := cmd.UserPermissionLevel(ctx)
	if err != nil {
		sentry.ErrorWithContext(err, cmd.ToErrorContext())
		return true, 1 // TODO: Stop flow
	}

	if isStaff >= permcache.Support {
		return false, 50
	}

	var openedTickets []database.Ticket
	var ticketLimit uint8

	group, _ := errgroup.WithContext(ctx)

	// get ticket limit
	group.Go(func() (err error) {
		ticketLimit, err = dbclient.Client.TicketLimit.Get(ctx, cmd.GuildId())
		return
	})

	group.Go(func() (err error) {
		openedTickets, err = dbclient.Client.Tickets.GetOpenByUser(ctx, cmd.GuildId(), cmd.UserId())
		return
	})

	if err := group.Wait(); err != nil {
		sentry.ErrorWithContext(err, cmd.ToErrorContext())
		return true, 1
	}

	return len(openedTickets) >= int(ticketLimit), int(ticketLimit)
}

func createWebhook(ctx context.Context, c registry.CommandContext, ticketId int, guildId, channelId uint64) error {
	// TODO: Re-add permission check
	//if permission.HasPermissionsChannel(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), channelId, permission.ManageWebhooks) { // Do we actually need this?
	root := sentry.StartSpan(ctx, "Create webhook")
	defer root.Finish()

	span := sentry.StartSpan(root.Context(), "Get bot user")
	self, err := c.Worker().Self()
	span.Finish()
	if err != nil {
		return err
	}

	data := rest.WebhookData{
		Username: self.Username,
		Avatar:   self.AvatarUrl(256),
	}

	span = sentry.StartSpan(root.Context(), "Get bot user")
	webhook, err := c.Worker().CreateWebhook(channelId, data)
	span.Finish()
	if err != nil {
		sentry.ErrorWithContext(err, c.ToErrorContext())
		return nil // Silently fail
	}

	dbWebhook := database.Webhook{
		Id:    webhook.Id,
		Token: webhook.Token,
	}

	span = sentry.StartSpan(root.Context(), "Store webhook in database")
	defer span.Finish()
	if err := dbclient.Client.Webhooks.Create(ctx, guildId, ticketId, dbWebhook); err != nil {
		return err
	}

	return nil
}

func CreateOverwrites(ctx context.Context, cmd registry.InteractionContext, userId uint64, panel *database.Panel, otherUsers ...uint64) ([]channel.PermissionOverwrite, error) {
	overwrites := []channel.PermissionOverwrite{ // @everyone
		{
			Id:    cmd.GuildId(),
			Type:  channel.PermissionTypeRole,
			Allow: 0,
			Deny:  permission.BuildPermissions(permission.ViewChannel),
		},
	}

	// Build permissions
	additionalPermissions, err := dbclient.Client.TicketPermissions.Get(ctx, cmd.GuildId())
	if err != nil {
		return nil, err
	}

	// Separate permissions apply
	for _, snowflake := range append(otherUsers, userId) {
		overwrites = append(overwrites, BuildUserOverwrite(snowflake, additionalPermissions))
	}

	// Add the bot to the overwrites
	selfAllow := make([]permission.Permission, len(StandardPermissions), len(StandardPermissions)+1)
	copy(selfAllow, StandardPermissions[:]) // Do not append to StandardPermissions

	if permission.HasPermissionRaw(cmd.InteractionMetadata().AppPermissions, permission.ManageWebhooks) {
		selfAllow = append(selfAllow, permission.ManageWebhooks)
	}

	integrationRoleId, err := GetIntegrationRoleId(ctx, cmd.Worker(), cmd.GuildId())
	if err != nil {
		return nil, err
	}

	if integrationRoleId == nil {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    cmd.Worker().BotId,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(selfAllow[:]...),
			Deny:  0,
		})
	} else {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    *integrationRoleId,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(selfAllow[:]...),
			Deny:  0,
		})
	}

	// Create list of members & roles who should be added to the ticket
	allowedUsers, allowedRoles, err := GetAllowedStaffUsersAndRoles(ctx, cmd.GuildId(), panel)
	if err != nil {
		return nil, err
	}

	for _, member := range allowedUsers {
		allow := make([]permission.Permission, len(StandardPermissions))
		copy(allow, StandardPermissions[:]) // Do not append to StandardPermissions

		if member == cmd.Worker().BotId {
			continue // Already added overwrite above
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
			Allow: permission.BuildPermissions(StandardPermissions[:]...),
			Deny:  0,
		})
	}

	return overwrites, nil
}

func GetAllowedStaffUsersAndRoles(ctx context.Context, guildId uint64, panel *database.Panel) ([]uint64, []uint64, error) {
	// Create list of members & roles who should be added to the ticket
	// Add the sender & self
	allowedUsers := make([]uint64, 0)
	allowedRoles := make([]uint64, 0)

	// Should we add the default team
	if panel == nil || panel.WithDefaultTeam {
		// Get support reps & admins
		supportUsers, err := dbclient.Client.Permissions.GetSupport(ctx, guildId)
		if err != nil {
			return nil, nil, err
		}

		allowedUsers = append(allowedUsers, supportUsers...)

		// Get support roles & admin roles
		supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(ctx, guildId)
		if err != nil {
			return nil, nil, err
		}

		allowedRoles = append(allowedUsers, supportRoles...)
	}

	// Add other support teams
	if panel != nil {
		group, _ := errgroup.WithContext(ctx)

		// Get users for support teams of panel
		group.Go(func() error {
			userIds, err := dbclient.Client.SupportTeamMembers.GetAllSupportMembersForPanel(ctx, panel.PanelId)
			if err != nil {
				return err
			}

			allowedUsers = append(allowedUsers, userIds...) // No mutex needed
			return nil
		})

		// Get roles for support teams of panel
		group.Go(func() error {
			roleIds, err := dbclient.Client.SupportTeamRoles.GetAllSupportRolesForPanel(ctx, panel.PanelId)
			if err != nil {
				return err
			}

			allowedRoles = append(allowedRoles, roleIds...) // No mutex needed
			return nil
		})

		if err := group.Wait(); err != nil {
			return nil, nil, err
		}
	}

	return allowedUsers, allowedRoles, nil
}

func GetIntegrationRoleId(rootCtx context.Context, worker *worker.Context, guildId uint64) (*uint64, error) {
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*3)
	defer cancel()

	cachedId, err := redis.GetIntegrationRole(ctx, guildId, worker.BotId)
	if err == nil {
		return &cachedId, nil
	} else if !errors.Is(err, redis.ErrIntegrationRoleNotCached) {
		return nil, err
	}

	roles, err := worker.GetGuildRoles(guildId)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		if role.Tags.BotId != nil && *role.Tags.BotId == worker.BotId {
			ctx, cancel := context.WithTimeout(rootCtx, time.Second*3)
			defer cancel() // defer is okay here as we return in every case

			if err := redis.SetIntegrationRole(ctx, guildId, worker.BotId, role.Id); err != nil {
				return nil, err
			}

			return &role.Id, nil
		}
	}

	return nil, nil
}

func GenerateChannelName(ctx context.Context, cmd registry.CommandContext, panel *database.Panel, ticketId int, openerId uint64, claimer *uint64) (string, error) {
	// Create ticket name
	var name string

	// Use server default naming scheme
	if panel == nil || panel.NamingScheme == nil {
		namingScheme, err := dbclient.Client.NamingScheme.Get(ctx, cmd.GuildId())
		if err != nil {
			return "", err
		}

		strTicket := strings.ToLower(cmd.GetMessage(i18n.Ticket))
		if namingScheme == database.Username {
			var user user.User
			if cmd.UserId() == openerId {
				user, err = cmd.User()
			} else {
				user, err = cmd.Worker().GetUser(openerId)
			}

			if err != nil {
				return "", err
			}

			name = fmt.Sprintf("%s-%s", strTicket, user.Username)
		} else {
			name = fmt.Sprintf("%s-%d", strTicket, ticketId)
		}
	} else {
		var err error
		name, err = doSubstitutions(cmd, *panel.NamingScheme, openerId, []Substitutor{
			// %id%
			NewSubstitutor("id", false, false, func(user user.User, member member.Member) string {
				return strconv.Itoa(ticketId)
			}),
			// %id_padded%
			NewSubstitutor("id_padded", false, false, func(user user.User, member member.Member) string {
				return fmt.Sprintf("%04d", ticketId)
			}),
			// %claimed%
			NewSubstitutor("claimed", false, false, func(user user.User, member member.Member) string {
				if claimer == nil {
					return "unclaimed"
				} else {
					return "claimed"
				}
			}),
			// %username%
			NewSubstitutor("username", true, false, func(user user.User, member member.Member) string {
				return user.Username
			}),
			// %nickname%
			NewSubstitutor("nickname", false, true, func(user user.User, member member.Member) string {
				nickname := member.Nick
				if len(nickname) == 0 {
					nickname = member.User.Username
				}

				return nickname
			}),
		})

		if err != nil {
			return "", err
		}
	}

	// Cap length after substitutions
	if len(name) > 100 {
		name = name[:100]
	}

	return name, nil
}

func countRealChannels(channels []channel.Channel, parentId uint64) int {
	var count int

	for _, ch := range channels {
		// Ignore threads
		if ch.Type == channel.ChannelTypeGuildPublicThread || ch.Type == channel.ChannelTypeGuildPrivateThread || ch.Type == channel.ChannelTypeGuildNewsThread {
			continue
		}

		if parentId == 0 || ch.ParentId.Value == parentId {
			count++
		}
	}

	return count
}

func BuildJoinThreadMessage(
	ctx context.Context,
	worker *worker.Context,
	guildId, openerId uint64,
	ticketId int,
	panel *database.Panel,
	staffMembers []uint64,
	premiumTier premium.PremiumTier,
) command.MessageResponse {
	return buildJoinThreadMessage(ctx, worker, guildId, openerId, ticketId, panel, staffMembers, premiumTier, false)
}

func BuildThreadReopenMessage(
	ctx context.Context,
	worker *worker.Context,
	guildId, openerId uint64,
	ticketId int,
	panel *database.Panel,
	staffMembers []uint64,
	premiumTier premium.PremiumTier,
) command.MessageResponse {
	return buildJoinThreadMessage(ctx, worker, guildId, openerId, ticketId, panel, staffMembers, premiumTier, true)
}

// TODO: Translations
func buildJoinThreadMessage(
	ctx context.Context,
	worker *worker.Context,
	guildId, openerId uint64,
	ticketId int,
	panel *database.Panel,
	staffMembers []uint64,
	premiumTier premium.PremiumTier,
	fromReopen bool,
) command.MessageResponse {
	var colour customisation.Colour
	if len(staffMembers) > 0 {
		colour = customisation.Green
	} else {
		colour = customisation.Red
	}

	panelName := "None"
	if panel != nil {
		panelName = panel.ButtonLabel
	}

	title := "Join Ticket"
	if fromReopen {
		title = "Ticket Reopened"
	}

	e := utils.BuildEmbedRaw(customisation.GetColourOrDefault(ctx, guildId, colour), title, "A ticket has been opened. Press the button below to join it.", nil, premiumTier)
	e.AddField(customisation.PrefixWithEmoji("Opened By", customisation.EmojiOpen, !worker.IsWhitelabel), customisation.PrefixWithEmoji(fmt.Sprintf("<@%d>", openerId), customisation.EmojiBulletLine, !worker.IsWhitelabel), true)
	e.AddField(customisation.PrefixWithEmoji("Panel", customisation.EmojiPanel, !worker.IsWhitelabel), customisation.PrefixWithEmoji(panelName, customisation.EmojiBulletLine, !worker.IsWhitelabel), true)
	e.AddField(customisation.PrefixWithEmoji("Staff In Ticket", customisation.EmojiStaff, !worker.IsWhitelabel), customisation.PrefixWithEmoji(strconv.Itoa(len(staffMembers)), customisation.EmojiBulletLine, !worker.IsWhitelabel), true)

	if len(staffMembers) > 0 {
		var mentions []string // dynamic length
		charCount := len(customisation.EmojiBulletLine.String()) + 1
		for _, staffMember := range staffMembers {
			mention := fmt.Sprintf("<@%d>", staffMember)

			if charCount+len(mention)+1 > 1024 {
				break
			}

			mentions = append(mentions, mention)
			charCount += len(mention) + 1 // +1 for space
		}

		e.AddField(customisation.PrefixWithEmoji("Staff Members", customisation.EmojiStaff, !worker.IsWhitelabel), customisation.PrefixWithEmoji(strings.Join(mentions, " "), customisation.EmojiBulletLine, !worker.IsWhitelabel), false)
	}

	return command.MessageResponse{
		Embeds: utils.Slice(e),
		Components: utils.Slice(component.BuildActionRow(
			component.BuildButton(component.Button{
				Label:    "Join Ticket",
				CustomId: fmt.Sprintf("join_thread_%d", ticketId),
				Style:    component.ButtonStylePrimary,
				Emoji:    utils.BuildEmoji("➕"),
			}),
		)),
	}
}

func sendAccessControlDeniedMessage(ctx context.Context, cmd registry.InteractionContext, panelId int, matchedRole uint64) error {
	rules, err := dbclient.Client.PanelAccessControlRules.GetAll(ctx, panelId)
	if err != nil {
		return err
	}

	allowedRoleIds := make([]uint64, 0, len(rules))
	for _, rule := range rules {
		if rule.Action == database.AccessControlActionAllow {
			allowedRoleIds = append(allowedRoleIds, rule.RoleId)
		}
	}

	if len(allowedRoleIds) == 0 {
		cmd.Reply(customisation.Red, i18n.MessageNoPermission, i18n.MessageOpenAclNoAllowRules)
		return nil
	}

	if matchedRole == cmd.GuildId() {
		mentions := make([]string, 0, len(allowedRoleIds))
		for _, roleId := range allowedRoleIds {
			mentions = append(mentions, fmt.Sprintf("<@&%d>", roleId))
		}

		if len(allowedRoleIds) == 1 {
			cmd.Reply(customisation.Red, i18n.MessageNoPermission, i18n.MessageOpenAclNotAllowListedSingle, strings.Join(mentions, ", "))
		} else {
			cmd.Reply(customisation.Red, i18n.MessageNoPermission, i18n.MessageOpenAclNotAllowListedMultiple, strings.Join(mentions, ", "))
		}
	} else {
		cmd.Reply(customisation.Red, i18n.MessageNoPermission, i18n.MessageOpenAclDenyListed, matchedRole)
	}

	return nil
}
