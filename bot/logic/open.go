package logic

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
	"strings"
)

// if panel != nil, msg should be artifically filled, excluding the message ID
func OpenTicket(worker *worker.Context, user user.User, guildId, channelId, messageId uint64, isPremium bool, args []string, panel *database.Panel) {
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   guildId,
		User:    user.Id,
		Channel: channelId,
		Shard:   worker.ShardId,
		Command: "open",
	}

	// If we're using a panel, then we need to create the ticket in the specified category
	var category uint64
	if panel != nil && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		var err error
		category, err = dbclient.Client.ChannelCategory.Get(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
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
		ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
		if ctx.ShouldReact {
			ctx.ReactWithCross()
		}
		return
	}*/

	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		_, err := worker.GetChannel(category)
		if err != nil {
			useCategory = false
			//go database.DeleteCategory(ctx.GuildId) TODO: Could this be due to a Discord outage? Check specifically for a 404
		} else {
			// TODO: Re-add permission check
			/*if !permission.HasPermissionsChannel(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), category, requiredPerms...) {
				ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
				if ctx.ShouldReact {
					ctx.ReactWithCross()
				}
				return
			}*/
		}
	}

	// create DM channel
	dmChannel, err := worker.CreateDM(user.Id)

	// target channel for messaging the user
	// either DMs or the channel where the command was run
	var targetChannel uint64
	if panel == nil {
		targetChannel = channelId
	} else {
		targetChannel = dmChannel.Id
	}

	// Make sure ticket count is within ticket limit
	violatesTicketLimit, limit := getTicketLimit(guildId, user.Id)
	if violatesTicketLimit {
		// Notify the user
		if targetChannel != 0 {
			ticketsPluralised := "ticket"
			if limit > 1 {
				ticketsPluralised += "s"
			}

			utils.SendEmbed(worker, targetChannel, guildId, utils.Red, "Error", translations.MessageTicketLimitReached, nil, 30, isPremium, limit, ticketsPluralised)
		}

		return
	}

	// Generate subject
	subject := "No subject given"
	if panel != nil && panel.Title != "" { // If we're using a panel, use the panel title as the subject
		subject = panel.Title
	} else { // Else, take command args as the subject
		if len(args) > 0 {
			subject = strings.Join(args, " ")
		}
		if len(subject) > 256 {
			subject = subject[0:255]
		}
	}

	// Make sure there's not > 50 channels in a category
	if useCategory {
		channels, _ := worker.GetGuildChannels(guildId)

		channelCount := 0
		for _, channel := range channels {
			if channel.ParentId == category {
				channelCount++
			}
		}

		if channelCount >= 50 {
			utils.SendEmbed(worker, channelId, guildId, utils.Red, "Error", translations.MessageTooManyTickets, nil, 30, isPremium)
			return
		}
	}

	if panel == nil {
		utils.ReactWithCheck(worker, channelId, messageId)
	}

	// Create channel
	id, err := dbclient.Client.Tickets.Create(guildId, user.Id)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	overwrites := CreateOverwrites(guildId, user.Id, worker.BotId)

	// Create ticket name
	var name string

	namingScheme, err := dbclient.Client.NamingScheme.Get(guildId)
	if err != nil {
		namingScheme = database.Id
		sentry.ErrorWithContext(err, errorContext)
	}

	if namingScheme == database.Username {
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

	channel, err := worker.CreateGuildChannel(guildId, data)
	if err != nil { // Bot likely doesn't have permission
		sentry.ErrorWithContext(err, errorContext)

		// To prevent tickets getting in a glitched state, we should mark it as closed (or delete it completely?)
		if err := dbclient.Client.Tickets.Close(id, guildId); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}

		return
	}

	welcomeMessageId, err := utils.SendWelcomeMessage(worker, guildId, channel.Id, user.Id, isPremium, subject, panel, id)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	// UpdateUser channel in DB
	go func() {
		if err := dbclient.Client.Tickets.SetTicketProperties(guildId, id, channel.Id, welcomeMessageId); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}()

	// mentions
	{
		var content string

		if panel == nil {
			// Ping @everyone
			pingEveryone, err := dbclient.Client.PingEveryone.Get(guildId)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}

			if pingEveryone {
				content = fmt.Sprintf("@everyone")
			}
		} else {
			// roles
			roles, err := dbclient.Client.PanelRoleMentions.GetRoles(panel.MessageId)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
			} else {
				for _, roleId := range roles {
					content += fmt.Sprintf("<@&%d>", roleId)
				}
			}

			// user
			shouldMentionUser, err := dbclient.Client.PanelUserMention.ShouldMentionUser(panel.MessageId)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
			} else {
				if shouldMentionUser {
					content += fmt.Sprintf("<@%d>", user.Id)
				}
			}
		}

		if content != "" {
			if len(content) > 2000 {
				content = content[:2000]
			}

			pingMessage, err := worker.CreateMessageComplex(channel.Id, rest.CreateMessageData{
				Content:         content,
				AllowedMentions: message.AllowedMention{
					Parse: []message.AllowedMentionType{
						message.EVERYONE,
						message.USERS,
						message.ROLES,
					},
				},
			})

			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
			} else {
				// error is likely to be a permission error
				_ = worker.DeleteMessage(channel.Id, pingMessage.Id)
			}
		}
	}

	// Let the user know the ticket has been opened
	if panel == nil {
		utils.SendEmbed(worker, channelId, guildId, utils.Green, "Ticket", translations.MessageTicketOpened, nil, 30, isPremium, channel.Mention())
	} else {
		dmOnOpen, err := dbclient.Client.DmOnOpen.Get(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}

		if dmOnOpen && dmChannel.Id != 0 {
			utils.SendEmbed(worker, dmChannel.Id, guildId, utils.Green, "Ticket", translations.MessageTicketOpened, nil, 0, isPremium, channel.Mention())
		}
	}

	go statsd.Client.IncrementKey(statsd.TICKETS)

	if isPremium {
		go createWebhook(worker, id, guildId, channel.Id)
	}

	// update cache
	go func() {
		// retrieve member
		// GetGuildMember will cache if not already cached
		if _, err := worker.GetGuildMember(guildId, user.Id); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}

		// store user
		cache.Client.StoreUser(user)
	}()
}

// has hit ticket limit, ticket limit
func getTicketLimit(guildId, userId uint64) (bool, int) {
	var openedTickets []database.Ticket
	var ticketLimit uint8

	group, _ := errgroup.WithContext(context.Background())

	// get ticket limit
	group.Go(func() (err error) {
		ticketLimit, err = dbclient.Client.TicketLimit.Get(guildId)
		return
	})

	group.Go(func() (err error) {
		openedTickets, err = dbclient.Client.Tickets.GetOpenByUser(guildId, userId)
		return
	})

	if err := group.Wait(); err != nil {
		sentry.Error(err)
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
	//}
}

func CreateOverwrites(guildId, userId, selfId uint64) (overwrites []channel.PermissionOverwrite) {
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   guildId,
		User:    userId,
		Command: "open",
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

	// Get support reps & admins
	supportUsers, err := dbclient.Client.Permissions.GetSupport(guildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	for _, user := range supportUsers {
		allowedUsers = append(allowedUsers, user)
	}

	// Get support roles & admin roles
	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(guildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	for _, role := range supportRoles {
		allowedRoles = append(allowedRoles, role)
	}

	// Add ourselves and the sender
	allowedUsers = append(allowedUsers, selfId, userId)

	for _, member := range allowedUsers {
		allow := []permission.Permission{permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks}

		// Give ourselves permissions to create webbooks
		if member == selfId {
			allow = append(allow, permission.ManageWebhooks)
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
			Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
			Deny:  0,
		})
	}

	return overwrites
}
