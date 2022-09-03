package logic

import (
	"context"
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
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/permissionwrapper"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	model "github.com/rxdn/gdl/objects/guild"
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

func OpenTicket(ctx registry.CommandContext, panel *database.Panel, subject string, formData map[database.FormInput]string) (database.Ticket, error) {
	// Make sure ticket count is within ticket limit
	// Check ticket limit before ratelimit token to prevent 1 person from stopping everyone opening tickets
	violatesTicketLimit, limit := getTicketLimit(ctx)
	if violatesTicketLimit {
		// Notify the user
		ticketsPluralised := "ticket"
		if limit > 1 {
			ticketsPluralised += "s"
		}

		// TODO: Use translation of tickets
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageTicketLimitReached, limit, ticketsPluralised)
		return database.Ticket{}, fmt.Errorf("ticket limit reached")
	}

	ok, err := redis.TakeTicketRateLimitToken(redis.Client, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return database.Ticket{}, err
	}

	if !ok {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageOpenRatelimited)
		return database.Ticket{}, nil
	}

	settings, err := dbclient.Client.Settings.Get(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return database.Ticket{}, err
	}

	guild, err := ctx.Guild()
	if err != nil {
		ctx.HandleError(err)
		return database.Ticket{}, err
	}

	isThread := settings.UseThreads && guild.PremiumTier >= model.PremiumTier2

	// If we're using a panel, then we need to create the ticket in the specified category
	var category uint64
	if panel != nil && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		var err error
		category, err = dbclient.Client.ChannelCategory.Get(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return database.Ticket{}, err
		}
	}

	useCategory := category != 0 && !isThread
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
		}
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
	channels, _ := ctx.Worker().GetGuildChannels(ctx.GuildId())

	// 500 guild limit check
	if !isThread && countRealChannels(channels, 0) >= 500 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageGuildChannelLimitReached)
		return database.Ticket{}, fmt.Errorf("channel limit reached")
	}

	// Make sure there's not > 50 channels in a category
	if useCategory {
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
					if _, err := ctx.Worker().GetChannel(category); err != nil {
						if restError, ok := err.(request.RestError); ok && restError.StatusCode == 404 {
							if err := dbclient.Client.Settings.SetOverflow(ctx.GuildId(), false, nil); err != nil {
								ctx.HandleError(err)
								return database.Ticket{}, err
							}
						}

						ctx.Reply(customisation.Red, i18n.Error, i18n.MessageTooManyTickets)
						return database.Ticket{}, err
					}

					// Check that the overflow category still has space
					overflowCategoryChildrenCount := countRealChannels(channels, *settings.OverflowCategoryId)

					if overflowCategoryChildrenCount >= 50 {
						ctx.Reply(customisation.Red, i18n.Error, i18n.MessageTooManyTickets)
						return database.Ticket{}, fmt.Errorf("overflow category full")
					}
				}
			} else {
				ctx.Reply(customisation.Red, i18n.Error, i18n.MessageTooManyTickets)
				return database.Ticket{}, fmt.Errorf("category ticket limit reached")
			}
		}
	}

	// Create channel
	ticketId, err := dbclient.Client.Tickets.Create(ctx.GuildId(), ctx.UserId(), isThread)
	if err != nil {
		ctx.HandleError(err)
		return database.Ticket{}, err
	}

	name, err := GenerateChannelName(ctx, panel, ticketId, ctx.UserId(), nil)
	if err != nil {
		ctx.HandleError(err)
		return database.Ticket{}, err
	}

	var ch channel.Channel
	var joinMessageId *uint64
	if isThread {
		ch, err = ctx.Worker().CreatePrivateThread(ctx.ChannelId(), name, uint16(settings.ThreadArchiveDuration), false)
		if err != nil {
			ctx.HandleError(err)

			// To prevent tickets getting in a glitched state, we should mark it as closed (or delete it completely?)
			if err := dbclient.Client.Tickets.Close(ticketId, ctx.GuildId()); err != nil {
				ctx.HandleError(err)
			}

			return database.Ticket{}, err
		}

		// Join ticket
		if err := ctx.Worker().AddThreadMember(ch.Id, ctx.UserId()); err != nil {
			ctx.HandleError(err)
		}

		if settings.TicketNotificationChannel != nil {
			data := BuildJoinThreadMessage(ctx.Worker(), ctx.GuildId(), ctx.UserId(), ticketId, panel, nil, ctx.PremiumTier())

			if msg, err := ctx.Worker().CreateMessageComplex(*settings.TicketNotificationChannel, data.IntoCreateMessageData()); err == nil {
				joinMessageId = &msg.Id
			} else {
				ctx.HandleError(err)
			}
		}
	} else {
		overwrites, err := CreateOverwrites(ctx.Worker(), ctx.GuildId(), ctx.UserId(), ctx.Worker().BotId, panel)
		if err != nil {
			ctx.HandleError(err)
			return database.Ticket{}, err
		}

		data := rest.CreateChannelData{
			Name:                 name,
			Type:                 channel.ChannelTypeGuildText,
			Topic:                subject,
			PermissionOverwrites: overwrites,
		}

		if useCategory {
			data.ParentId = category
		}

		tmp, err := ctx.Worker().CreateGuildChannel(ctx.GuildId(), data)
		if err != nil { // Bot likely doesn't have permission
			ctx.HandleError(err)

			// To prevent tickets getting in a glitched state, we should mark it as closed (or delete it completely?)
			if err := dbclient.Client.Tickets.Close(ticketId, ctx.GuildId()); err != nil {
				ctx.HandleError(err)
			}

			return database.Ticket{}, err
		}

		// TODO: Remove
		if tmp.Id == 0 {
			ctx.HandleError(fmt.Errorf("channel id is 0"))
			return database.Ticket{}, fmt.Errorf("channel id is 0")
		}

		ch = tmp
	}

	ctx.Accept()

	var panelId *int
	if panel != nil {
		panelId = &panel.PanelId
	}

	ticket := database.Ticket{
		Id:               ticketId,
		GuildId:          ctx.GuildId(),
		ChannelId:        &ch.Id,
		UserId:           ctx.UserId(),
		Open:             true,
		OpenTime:         time.Now(), // will be a bit off, but not used
		WelcomeMessageId: nil,
		PanelId:          panelId,
		IsThread:         isThread,
		JoinMessageId:    joinMessageId,
	}

	welcomeMessageId, err := SendWelcomeMessage(ctx, ticket, subject, panel, formData)
	if err != nil {
		ctx.HandleError(err)
	}

	if welcomeMessageId != 0 {
		ticket.WelcomeMessageId = &welcomeMessageId
	}

	// UpdateUser channel in DB
	if err := dbclient.Client.Tickets.SetTicketProperties(ctx.GuildId(), ticketId, ch.Id, welcomeMessageId, joinMessageId, panelId); err != nil {
		ctx.HandleError(err)
	}

	metadata, err := dbclient.Client.GuildMetadata.Get(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return database.Ticket{}, err
	}

	// mentions
	{
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

				teams, err := dbclient.Client.PanelTeams.GetTeams(panel.PanelId)
				if err != nil {
					ctx.HandleError(err)
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

			pingMessage, err := ctx.Worker().CreateMessageComplex(ch.Id, rest.CreateMessageData{
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
				_ = ctx.Worker().DeleteMessage(ch.Id, pingMessage.Id)
			}
		}
	}

	// Let the user know the ticket has been opened
	// Ephemeral reply is ok
	ctx.Reply(customisation.Green, i18n.Ticket, i18n.MessageTicketOpened, ch.Mention())

	prometheus.LogTicketCreated(ctx.GuildId())
	statsd.Client.IncrementKey(statsd.KeyTickets)
	if panel == nil {
		statsd.Client.IncrementKey(statsd.KeyOpenCommand)
	}

	if ctx.PremiumTier() > premium.None {
		go createWebhook(ctx.Worker(), ticketId, ctx.GuildId(), ch.Id)
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

	return ticket, nil
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

func CreateOverwrites(worker *worker.Context, guildId, userId, selfId uint64, panel *database.Panel, otherUsers ...uint64) ([]channel.PermissionOverwrite, error) {
	overwrites := []channel.PermissionOverwrite{ // @everyone
		{
			Id:    guildId,
			Type:  channel.PermissionTypeRole,
			Allow: 0,
			Deny:  permission.BuildPermissions(permission.ViewChannel),
		},
	}

	// Build permissions
	additionalPermissions, err := dbclient.Client.TicketPermissions.Get(guildId)
	if err != nil {
		return nil, err
	}

	// Separate permissions apply
	for _, snowflake := range append(otherUsers, userId) {
		overwrites = append(overwrites, BuildUserOverwrite(snowflake, additionalPermissions))
	}

	// Create list of members & roles who should be added to the ticket
	allowedUsers, allowedRoles, err := getAllowedUsersRoles(guildId, selfId, panel)
	if err != nil {
		return nil, err
	}

	for _, member := range allowedUsers {
		allow := make([]permission.Permission, len(StandardPermissions))
		copy(allow, StandardPermissions[:]) // Do not append to StandardPermissions

		// Give ourselves permissions to create webhooks
		if member == selfId {
			if permissionwrapper.HasPermissions(worker, guildId, selfId, permission.ManageWebhooks) {
				allow = append(allow, permission.ManageWebhooks)
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
			Allow: permission.BuildPermissions(StandardPermissions[:]...),
			Deny:  0,
		})
	}

	return overwrites, nil
}

func getAllowedUsersRoles(guildId, selfId uint64, panel *database.Panel) ([]uint64, []uint64, error) {
	errorContext := errorcontext.WorkerErrorContext{
		Guild: guildId,
	}

	// Create list of members & roles who should be added to the ticket
	// Add the sender & self
	allowedUsers := []uint64{selfId}
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
		group, _ := errgroup.WithContext(context.Background())

		// Get users for support teams of panel
		group.Go(func() error {
			userIds, err := dbclient.Client.SupportTeamMembers.GetAllSupportMembersForPanel(panel.PanelId)
			if err != nil {
				return err
			}

			allowedUsers = append(allowedUsers, userIds...) // No mutex needed
			return nil
		})

		// Get roles for support teams of panel
		group.Go(func() error {
			roleIds, err := dbclient.Client.SupportTeamRoles.GetAllSupportRolesForPanel(panel.PanelId)
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

func GenerateChannelName(ctx registry.CommandContext, panel *database.Panel, ticketId int, openerId uint64, claimer *uint64) (string, error) {
	// Create ticket name
	var name string

	// Use server default naming scheme
	if panel == nil || panel.NamingScheme == nil {
		namingScheme, err := dbclient.Client.NamingScheme.Get(ctx.GuildId())
		if err != nil {
			return "", err
		}

		strTicket := strings.ToLower(ctx.GetMessage(i18n.Ticket))
		if namingScheme == database.Username {
			var user user.User
			if ctx.UserId() == openerId {
				user, err = ctx.User()
			} else {
				user, err = ctx.Worker().GetUser(openerId)
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
		name, err = doSubstitutions(ctx, *panel.NamingScheme, openerId, []Substitutor{
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
	worker *worker.Context,
	guildId, openerId uint64,
	ticketId int,
	panel *database.Panel,
	staffMembers []uint64,
	premiumTier premium.PremiumTier,
) command.MessageResponse {
	return buildJoinThreadMessage(worker, guildId, openerId, ticketId, panel, staffMembers, premiumTier, false)
}

func BuildThreadReopenMessage(
	worker *worker.Context,
	guildId, openerId uint64,
	ticketId int,
	panel *database.Panel,
	staffMembers []uint64,
	premiumTier premium.PremiumTier,
) command.MessageResponse {
	return buildJoinThreadMessage(worker, guildId, openerId, ticketId, panel, staffMembers, premiumTier, true)
}

// TODO: Translations
func buildJoinThreadMessage(
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

	e := utils.BuildEmbedRaw(customisation.GetColourOrDefault(guildId, colour), title, "A ticket has been opened. Press the button below to join it.", nil, premiumTier)
	e.AddField(utils.PrefixWithEmoji("Opened By", utils.EmojiOpen, !worker.IsWhitelabel), utils.PrefixWithEmoji(fmt.Sprintf("<@%d>", openerId), utils.EmojiBulletLine, !worker.IsWhitelabel), true)
	e.AddField(utils.PrefixWithEmoji("Panel", utils.EmojiPanel, !worker.IsWhitelabel), utils.PrefixWithEmoji(panelName, utils.EmojiBulletLine, !worker.IsWhitelabel), true)
	e.AddField(utils.PrefixWithEmoji("Staff In Ticket", utils.EmojiStaff, !worker.IsWhitelabel), utils.PrefixWithEmoji(strconv.Itoa(len(staffMembers)), utils.EmojiBulletLine, !worker.IsWhitelabel), true)

	if len(staffMembers) > 0 {
		var mentions []string // dynamic length
		charCount := len(utils.EmojiBulletLine.String()) + 1
		for _, staffMember := range staffMembers {
			mention := fmt.Sprintf("<@%d>", staffMember)

			if charCount+len(mention)+1 > 1024 {
				break
			}

			mentions = append(mentions, mention)
			charCount += len(mention) + 1 // +1 for space
		}

		e.AddField(utils.PrefixWithEmoji("Staff Members", utils.EmojiStaff, !worker.IsWhitelabel), utils.PrefixWithEmoji(strings.Join(mentions, " "), utils.EmojiBulletLine, !worker.IsWhitelabel), false)
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
