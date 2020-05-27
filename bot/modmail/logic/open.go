package logic

import (
	"context"
	"errors"
	"fmt"
	permwrapper "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/gofrs/uuid"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
)

func OpenModMailTicket(worker *worker.Context, guild guild.Guild, user user.User, welcomeMessageId uint64) (uint64, error) {
	category, err := dbclient.Client.ChannelCategory.Get(guild.Id)
	if err != nil {
		sentry.Error(err)
	}

	// Make sure the category exists
	if category != 0 {
		if _, err := worker.GetChannel(category); err != nil {
			category = 0
		}
	}

	requiredPerms := []permission.Permission{
		permission.ManageChannels,
		permission.ManageRoles,
		permission.ViewChannel,
		permission.SendMessages,
		permission.ReadMessageHistory,
	}

	// TODO: Re-add permission check
	/*if !permission.HasPermissions(shard, guild.Id, shard.SelfId(), requiredPerms...) {
		return 0, errors.New("I do not have the correct permissions required to create the channel in the server")
	}*/

	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		// TODO: Decide whether to remove this check
		_, err := worker.GetChannel(category)
		if err != nil {
			useCategory = false
			go dbclient.Client.ChannelCategory.Delete(guild.Id)
			return 0, errors.New("Ticket category has been deleted")
		}

		if !permwrapper.HasPermissionsChannel(utils.ToRetriever(worker), guild.Id, worker.BotId, category, requiredPerms...) {
			return 0, errors.New("I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions")
		}
	}

	if useCategory {
		channels, err := worker.GetGuildChannels(guild.Id); if err != nil {
			return 0, err
		}

		channelCount := 0
		for _, channel := range channels {
			if channel.ParentId == category {
				channelCount += 1
			}
		}

		if channelCount >= 50 {
			return 0, errors.New("There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category")
		}
	}

	// Create channel
	name := fmt.Sprintf("modmail-%s", user.Username)
	overwrites := createOverwrites(worker, guild.Id)

	data := rest.CreateChannelData{
		Name:                 name,
		Type:                 channel.ChannelTypeGuildText,
		PermissionOverwrites: overwrites,
		ParentId:             category, // If not using category, value will be 0 and omitempty
	}

	channel, err := worker.CreateGuildChannel(guild.Id, data)
	if err != nil {
		sentry.Error(err)
		return 0, err
	}

	uuid, err := dbclient.Client.ModmailSession.Create(database.ModmailSession{
		GuildId:          guild.Id,
		UserId:           user.Id,
		StaffChannelId:   channel.Id,
		WelcomeMessageId: welcomeMessageId,
	})
	if err != nil {
		sentry.Error(err)
		return 0, err
	}

	// Create webhook
	go createWebhook(worker, guild.Id, channel.Id, uuid)

	return channel.Id, nil
}

func createWebhook(worker *worker.Context, guildId, channelId uint64, uuid uuid.UUID) {
	self, err := worker.Self()
	if err != nil {
		sentry.Error(errors.New("self is not cached"))
		return
	}

	/*if permission.HasPermissionsChannel(shard, guildId, channelId, self.Id, permission.ManageWebhooks) { // Do we actually need this?
	}*/

	webhook, err := worker.CreateWebhook(channelId, rest.WebhookData{
		Username: self.Username,
		Avatar:   self.AvatarUrl(256),
	})
	if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   guildId,
			Shard:   worker.ShardId,
			Command: "open",
		})
		return
	}

	dbWebhook := database.ModmailWebhook{
		Uuid:         uuid,
		WebhookId:    webhook.Id,
		WebhookToken: webhook.Token,
	}

	if err := dbclient.Client.ModmailWebhook.Create(dbWebhook); err != nil {
		sentry.Error(err)
	}
}

func createOverwrites(worker *worker.Context, guildId uint64) (overwrites []channel.PermissionOverwrite) {
	// Apply permission overwrites
	overwrites = append(overwrites, channel.PermissionOverwrite{ // @everyone
		Id:    guildId,
		Type:  channel.PermissionTypeRole,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel),
	})

	// Get support reps & roles
	var supportUsers []uint64
	var supportRoles []uint64

	group, _ := errgroup.WithContext(context.Background())

	group.Go(func() (err error) {
		supportUsers, err = dbclient.Client.Permissions.GetSupport(guildId)
		return
	})

	group.Go(func() (err error) {
		supportRoles, err = dbclient.Client.RolePermissions.GetSupportRoles(guildId)
		return
	})

	if err := group.Wait(); err != nil {
		sentry.Error(err)
	}

	// Create list of members & roles who should be added to the ticket
	allowedUsers := supportUsers
	allowedRoles := supportRoles

	// Add ourselves
	allowedUsers = append(allowedUsers, worker.BotId)

	for _, member := range allowedUsers {
		allow := []permission.Permission{permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks}

		// Give ourselves permissions to create webbooks
		if member == worker.BotId {
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

