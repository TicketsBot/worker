package worker

import (
	"context"
	"errors"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/objects/auditlog"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/integration"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/invite"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
)

func (ctx *Context) GetChannel(channelId uint64) (channel.Channel, error) {
	shouldCache := ctx.Cache.Options().Channels
	if shouldCache {
		cached, err := ctx.Cache.GetChannel(context.Background(), channelId)
		if err == nil {
			return cached, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return channel.Channel{}, err
		} // else, continue
	}

	channel, err := rest.GetChannel(context.Background(), ctx.Token, ctx.RateLimiter, channelId)

	if shouldCache && err == nil {
		go ctx.Cache.StoreChannel(context.Background(), channel)
	}

	return channel, err
}

func (ctx *Context) ModifyChannel(channelId uint64, data rest.ModifyChannelData) (channel.Channel, error) {
	channel, err := rest.ModifyChannel(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)

	if ctx.Cache.Options().Channels && err != nil {
		go ctx.Cache.StoreChannel(context.Background(), channel)
	}

	return channel, err
}

func (ctx *Context) DeleteChannel(channelId uint64) (channel.Channel, error) {
	return rest.DeleteChannel(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) GetChannelMessages(channelId uint64, options rest.GetChannelMessagesData) ([]message.Message, error) {
	return rest.GetChannelMessages(context.Background(), ctx.Token, ctx.RateLimiter, channelId, options)
}

func (ctx *Context) GetChannelMessage(channelId, messageId uint64) (message.Message, error) {
	return rest.GetChannelMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) CreateMessage(channelId uint64, content string) (message.Message, error) {
	return ctx.CreateMessageComplex(channelId, rest.CreateMessageData{
		Content: content,
	})
}

func (ctx *Context) CreateMessageReply(channelId uint64, content string, reference *message.MessageReference) (message.Message, error) {
	return ctx.CreateMessageComplex(channelId, rest.CreateMessageData{
		Content:          content,
		MessageReference: reference,
	})
}

func (ctx *Context) CreateMessageEmbed(channelId uint64, embed ...*embed.Embed) (message.Message, error) {
	return ctx.CreateMessageComplex(channelId, rest.CreateMessageData{
		Embeds: embed,
	})
}

func (ctx *Context) CreateMessageEmbedReply(channelId uint64, e *embed.Embed, reference *message.MessageReference) (message.Message, error) {
	return ctx.CreateMessageComplex(channelId, rest.CreateMessageData{
		Embeds:           []*embed.Embed{e},
		MessageReference: reference,
	})
}

func (ctx *Context) CreateMessageComplex(channelId uint64, data rest.CreateMessageData) (message.Message, error) {
	return rest.CreateMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) CreateReaction(channelId, messageId uint64, emoji string) error {
	return rest.CreateReaction(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId, emoji)
}

func (ctx *Context) DeleteOwnReaction(channelId, messageId uint64, emoji string) error {
	return rest.DeleteOwnReaction(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId, emoji)
}

func (ctx *Context) DeleteUserReaction(channelId, messageId, userId uint64, emoji string) error {
	return rest.DeleteUserReaction(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId, userId, emoji)
}

func (ctx *Context) GetReactions(channelId, messageId uint64, emoji string, options rest.GetReactionsData) ([]user.User, error) {
	return rest.GetReactions(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId, emoji, options)
}

func (ctx *Context) DeleteAllReactions(channelId, messageId uint64) error {
	return rest.DeleteAllReactions(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) DeleteAllReactionsEmoji(channelId, messageId uint64, emoji string) error {
	return rest.DeleteAllReactionsEmoji(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId, emoji)
}

func (ctx *Context) EditMessage(channelId, messageId uint64, data rest.EditMessageData) (message.Message, error) {
	return rest.EditMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId, data)
}

func (ctx *Context) DeleteMessage(channelId, messageId uint64) error {
	return rest.DeleteMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) BulkDeleteMessages(channelId uint64, messages []uint64) error {
	return rest.BulkDeleteMessages(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messages)
}

func (ctx *Context) EditChannelPermissions(channelId uint64, updated channel.PermissionOverwrite) error {
	return rest.EditChannelPermissions(context.Background(), ctx.Token, ctx.RateLimiter, channelId, updated)
}

func (ctx *Context) GetChannelInvites(channelId uint64) ([]invite.InviteMetadata, error) {
	return rest.GetChannelInvites(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) CreateChannelInvite(channelId uint64, data rest.CreateInviteData) (invite.Invite, error) {
	return rest.CreateChannelInvite(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) DeleteChannelPermissions(channelId, overwriteId uint64) error {
	return rest.DeleteChannelPermissions(context.Background(), ctx.Token, ctx.RateLimiter, channelId, overwriteId)
}

func (ctx *Context) TriggerTypingIndicator(channelId uint64) error {
	return rest.TriggerTypingIndicator(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) GetPinnedMessages(channelId uint64) ([]message.Message, error) {
	return rest.GetPinnedMessages(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) AddPinnedChannelMessage(channelId, messageId uint64) error {
	return rest.AddPinnedChannelMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) DeletePinnedChannelMessage(channelId, messageId uint64) error {
	return rest.DeletePinnedChannelMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) JoinThread(channelId uint64) error {
	return rest.JoinThread(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) AddThreadMember(channelId, userId uint64) error {
	return rest.AddThreadMember(context.Background(), ctx.Token, ctx.RateLimiter, channelId, userId)
}

func (ctx *Context) LeaveThread(channelId uint64) error {
	return rest.LeaveThread(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) RemoveThreadMember(channelId, userId uint64) error {
	return rest.RemoveThreadMember(context.Background(), ctx.Token, ctx.RateLimiter, channelId, userId)
}

func (ctx *Context) GetThreadMember(channelId, userId uint64) (channel.ThreadMember, error) {
	return rest.GetThreadMember(context.Background(), ctx.Token, ctx.RateLimiter, channelId, userId)
}

func (ctx *Context) ListThreadMembers(channelId uint64) ([]channel.ThreadMember, error) {
	return rest.ListThreadMembers(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) ListActiveThreads(channelId uint64) (rest.ThreadsResponse, error) {
	return rest.ListActiveThreads(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) ListPublicArchivedThreads(channelId uint64, data rest.ListThreadsData) (rest.ThreadsResponse, error) {
	return rest.ListPublicArchivedThreads(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) ListPrivateArchivedThreads(channelId uint64, data rest.ListThreadsData) (rest.ThreadsResponse, error) {
	return rest.ListPrivateArchivedThreads(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) ListJoinedPrivateArchivedThreads(channelId uint64, data rest.ListThreadsData) (rest.ThreadsResponse, error) {
	return rest.ListPrivateArchivedThreads(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) StartThreadWithMessage(channelId, messageId uint64, data rest.StartThreadWithMessageData) (channel.Channel, error) {
	return rest.StartThreadWithMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, messageId, data)
}

func (ctx *Context) StartThreadWithoutMessage(channelId, messageId uint64, data rest.StartThreadWithoutMessageData) (channel.Channel, error) {
	return rest.StartThreadWithoutMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) CreatePublicThread(channelId uint64, name string, autoArchiveDuration uint16) (channel.Channel, error) {
	data := rest.StartThreadWithoutMessageData{
		Name:                name,
		AutoArchiveDuration: autoArchiveDuration,
		Type:                channel.ChannelTypeGuildPublicThread,
	}

	return rest.StartThreadWithoutMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) CreatePrivateThread(channelId uint64, name string, autoArchiveDuration uint16, invitable bool) (channel.Channel, error) {
	data := rest.StartThreadWithoutMessageData{
		Name:                name,
		AutoArchiveDuration: autoArchiveDuration,
		Type:                channel.ChannelTypeGuildPrivateThread,
		Invitable:           invitable,
	}

	return rest.StartThreadWithoutMessage(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) ListGuildEmojis(guildId uint64) ([]emoji.Emoji, error) {
	shouldCacheEmoji := ctx.Cache.Options().Emojis
	shouldCacheGuild := ctx.Cache.Options().Guilds

	if shouldCacheEmoji && shouldCacheGuild {
		guild, err := ctx.Cache.GetGuild(context.Background(), guildId)
		if err == nil {
			return guild.Emojis, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return nil, err
		} // else, continue
	}

	emojis, err := rest.ListGuildEmojis(context.Background(), ctx.Token, ctx.RateLimiter, guildId)

	if shouldCacheEmoji && err == nil {
		go ctx.Cache.StoreEmojis(context.Background(), emojis, guildId)
	}

	return emojis, err
}

func (ctx *Context) GetGuildEmoji(guildId uint64, emojiId uint64) (emoji.Emoji, error) {
	shouldCache := ctx.Cache.Options().Emojis
	if shouldCache {
		e, err := ctx.Cache.GetEmoji(context.Background(), emojiId)
		if err == nil {
			return e, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return emoji.Emoji{}, err
		} // else, continue
	}

	emoji, err := rest.GetGuildEmoji(context.Background(), ctx.Token, ctx.RateLimiter, guildId, emojiId)

	if shouldCache && err == nil {
		go ctx.Cache.StoreEmoji(context.Background(), emoji, guildId)
	}

	return emoji, err
}

func (ctx *Context) CreateGuildEmoji(guildId uint64, data rest.CreateEmojiData) (emoji.Emoji, error) {
	return rest.CreateGuildEmoji(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

// updating Image is not permitted
func (ctx *Context) ModifyGuildEmoji(guildId, emojiId uint64, data rest.CreateEmojiData) (emoji.Emoji, error) {
	return rest.ModifyGuildEmoji(context.Background(), ctx.Token, ctx.RateLimiter, guildId, emojiId, data)
}

func (ctx *Context) CreateGuild(data rest.CreateGuildData) (guild.Guild, error) {
	return rest.CreateGuild(context.Background(), ctx.Token, data)
}

func (ctx *Context) GetGuild(guildId uint64) (guild.Guild, error) {
	shouldCache := ctx.Cache.Options().Guilds

	if shouldCache {
		cachedGuild, err := ctx.Cache.GetGuild(context.Background(), guildId)
		if err == nil {
			return cachedGuild, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return guild.Guild{}, err
		} // else, continue
	}

	guild, err := rest.GetGuild(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
	if err == nil {
		go ctx.Cache.StoreGuild(context.Background(), guild)
	}

	return guild, err
}

func (ctx *Context) GetGuildPreview(guildId uint64) (guild.GuildPreview, error) {
	return rest.GetGuildPreview(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) ModifyGuild(guildId uint64, data rest.ModifyGuildData) (guild.Guild, error) {
	return rest.ModifyGuild(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) DeleteGuild(guildId uint64) error {
	return rest.DeleteGuild(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetGuildChannels(guildId uint64) ([]channel.Channel, error) {
	shouldCache := ctx.Cache.Options().Guilds && ctx.Cache.Options().Channels

	if shouldCache {
		cached, err := ctx.Cache.GetGuildChannels(context.Background(), guildId)
		if err != nil && !errors.Is(err, cache.ErrNotFound) {
			return nil, err
		} else if err == nil && len(cached) > 0 { // either not cached (more likely), or guild has no channels
			return cached, nil
		} // else continue
	}

	channels, err := rest.GetGuildChannels(context.Background(), ctx.Token, ctx.RateLimiter, guildId)

	if shouldCache && err == nil {
		go ctx.Cache.ReplaceChannels(context.Background(), guildId, channels)
	}

	return channels, err
}

func (ctx *Context) CreateGuildChannel(guildId uint64, data rest.CreateChannelData) (channel.Channel, error) {
	return rest.CreateGuildChannel(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) ModifyGuildChannelPositions(guildId uint64, positions []rest.Position) error {
	return rest.ModifyGuildChannelPositions(context.Background(), ctx.Token, ctx.RateLimiter, guildId, positions)
}

func (ctx *Context) GetGuildMember(guildId, userId uint64) (member.Member, error) {
	cacheGuilds := ctx.Cache.Options().Guilds
	cacheUsers := ctx.Cache.Options().Users

	if cacheGuilds && cacheUsers {
		m, err := ctx.Cache.GetMember(context.Background(), guildId, userId)
		if err == nil {
			return m, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return member.Member{}, err
		} // else, continue
	}

	member, err := rest.GetGuildMember(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId)

	if cacheGuilds && err == nil {
		go ctx.Cache.StoreMember(context.Background(), member, guildId)
	}

	return member, err
}

func (ctx *Context) SearchGuildMembers(guildId uint64, data rest.SearchGuildMembersData) ([]member.Member, error) {
	members, err := rest.SearchGuildMembers(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
	if err == nil {
		go ctx.Cache.StoreMembers(context.Background(), members, guildId)
	}

	return members, err
}

func (ctx *Context) ListGuildMembers(guildId uint64, data rest.ListGuildMembersData) ([]member.Member, error) {
	members, err := rest.ListGuildMembers(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
	if err == nil {
		go ctx.Cache.StoreMembers(context.Background(), members, guildId)
	}

	return members, err
}

func (ctx *Context) ModifyGuildMember(guildId, userId uint64, data rest.ModifyGuildMemberData) error {
	return rest.ModifyGuildMember(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId, data)
}

func (ctx *Context) ModifyCurrentUserNick(guildId uint64, nick string) error {
	return rest.ModifyCurrentUserNick(context.Background(), ctx.Token, ctx.RateLimiter, guildId, nick)
}

func (ctx *Context) AddGuildMemberRole(guildId, userId, roleId uint64) error {
	return rest.AddGuildMemberRole(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId, roleId)
}

func (ctx *Context) RemoveGuildMemberRole(guildId, userId, roleId uint64) error {
	return rest.RemoveGuildMemberRole(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId, roleId)
}

func (ctx *Context) RemoveGuildMember(guildId, userId uint64) error {
	return rest.RemoveGuildMember(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId)
}

func (ctx *Context) GetGuildBans(guildId uint64, data rest.GetGuildBansData) ([]guild.Ban, error) {
	return rest.GetGuildBans(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) GetGuildBan(guildId, userId uint64) (guild.Ban, error) {
	return rest.GetGuildBan(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId)
}

func (ctx *Context) CreateGuildBan(guildId, userId uint64, data rest.CreateGuildBanData) error {
	return rest.CreateGuildBan(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId, data)
}

func (ctx *Context) RemoveGuildBan(guildId, userId uint64) error {
	return rest.RemoveGuildBan(context.Background(), ctx.Token, ctx.RateLimiter, guildId, userId)
}

func (ctx *Context) GetGuildRoles(guildId uint64) ([]guild.Role, error) {
	shouldCache := ctx.Cache.Options().Guilds && ctx.Cache.Options().Roles
	if shouldCache {
		cached, err := ctx.Cache.GetGuildRoles(context.Background(), guildId)
		if err != nil && !errors.Is(err, cache.ErrNotFound) {
			return nil, err
		} else if err == nil && len(cached) > 0 { // either not cached (more likely), or guild has no roles
			return cached, nil
		} // else continue
	}

	roles, err := rest.GetGuildRoles(context.Background(), ctx.Token, ctx.RateLimiter, guildId)

	if shouldCache && err == nil {
		go ctx.Cache.StoreRoles(context.Background(), roles, guildId)
	}

	return roles, err
}

func (ctx *Context) CreateGuildRole(guildId uint64, data rest.GuildRoleData) (guild.Role, error) {
	return rest.CreateGuildRole(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) ModifyGuildRolePositions(guildId uint64, positions []rest.Position) ([]guild.Role, error) {
	return rest.ModifyGuildRolePositions(context.Background(), ctx.Token, ctx.RateLimiter, guildId, positions)
}

func (ctx *Context) ModifyGuildRole(guildId, roleId uint64, data rest.GuildRoleData) (guild.Role, error) {
	return rest.ModifyGuildRole(context.Background(), ctx.Token, ctx.RateLimiter, guildId, roleId, data)
}

func (ctx *Context) DeleteGuildRole(guildId, roleId uint64) error {
	return rest.DeleteGuildRole(context.Background(), ctx.Token, ctx.RateLimiter, guildId, roleId)
}

func (ctx *Context) GetGuildPruneCount(guildId uint64, days int) (int, error) {
	return rest.GetGuildPruneCount(context.Background(), ctx.Token, ctx.RateLimiter, guildId, days)
}

// computePruneCount = whether 'pruned' is returned, discouraged for large guilds
func (ctx *Context) BeginGuildPrune(guildId uint64, days int, computePruneCount bool) error {
	return rest.BeginGuildPrune(context.Background(), ctx.Token, ctx.RateLimiter, guildId, days, computePruneCount)
}

func (ctx *Context) GetGuildVoiceRegions(guildId uint64) ([]guild.VoiceRegion, error) {
	return rest.GetGuildVoiceRegions(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetGuildInvites(guildId uint64) ([]invite.InviteMetadata, error) {
	return rest.GetGuildInvites(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetGuildIntegrations(guildId uint64) ([]integration.Integration, error) {
	return rest.GetGuildIntegrations(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) CreateGuildIntegration(guildId uint64, data rest.CreateIntegrationData) error {
	return rest.CreateGuildIntegration(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) ModifyGuildIntegration(guildId, integrationId uint64, data rest.ModifyIntegrationData) error {
	return rest.ModifyGuildIntegration(context.Background(), ctx.Token, ctx.RateLimiter, guildId, integrationId, data)
}

func (ctx *Context) DeleteGuildIntegration(guildId, integrationId uint64) error {
	return rest.DeleteGuildIntegration(context.Background(), ctx.Token, ctx.RateLimiter, guildId, integrationId)
}

func (ctx *Context) SyncGuildIntegration(guildId, integrationId uint64) error {
	return rest.SyncGuildIntegration(context.Background(), ctx.Token, ctx.RateLimiter, guildId, integrationId)
}

func (ctx *Context) GetGuildEmbed(guildId uint64) (guild.GuildWidget, error) {
	return rest.GetGuildWidget(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) ModifyGuildEmbed(guildId uint64, data guild.GuildEmbed) (guild.GuildEmbed, error) {
	return rest.ModifyGuildEmbed(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

// returns invite object with only "code" and "uses" fields
func (ctx *Context) GetGuildVanityUrl(guildId uint64) (invite.Invite, error) {
	return rest.GetGuildVanityURL(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetInvite(inviteCode string, withCounts bool) (invite.Invite, error) {
	return rest.GetInvite(context.Background(), ctx.Token, ctx.RateLimiter, inviteCode, withCounts)
}

func (ctx *Context) DeleteInvite(inviteCode string) (invite.Invite, error) {
	return rest.DeleteInvite(context.Background(), ctx.Token, ctx.RateLimiter, inviteCode)
}

func (ctx *Context) GetCurrentUser() (user.User, error) {
	cached, err := ctx.Cache.GetSelf(context.Background())
	if err == nil {
		return cached, nil
	} else if !errors.Is(err, cache.ErrNotFound) {
		return user.User{}, err
	} // else, continue

	self, err := rest.GetCurrentUser(context.Background(), ctx.Token, ctx.RateLimiter)

	if err == nil {
		go ctx.Cache.StoreSelf(context.Background(), self)
	}

	return self, err
}

func (ctx *Context) GetUser(userId uint64) (user.User, error) {
	shouldCache := ctx.Cache.Options().Users

	if shouldCache {
		cached, err := ctx.Cache.GetUser(context.Background(), userId)
		if err == nil {
			return cached, nil
		} else if !errors.Is(err, cache.ErrNotFound) {
			return user.User{}, err
		} // else, continue
	}

	user, err := rest.GetUser(context.Background(), ctx.Token, ctx.RateLimiter, userId)

	if shouldCache && err == nil {
		go ctx.Cache.StoreUser(context.Background(), user)
	}

	return user, err
}

func (ctx *Context) ModifyCurrentUser(data rest.ModifyUserData) (user.User, error) {
	return rest.ModifyCurrentUser(context.Background(), ctx.Token, ctx.RateLimiter, data)
}

func (ctx *Context) GetCurrentUserGuilds(data rest.CurrentUserGuildsData) ([]guild.Guild, error) {
	return rest.GetCurrentUserGuilds(context.Background(), ctx.Token, ctx.RateLimiter, data)
}

func (ctx *Context) LeaveGuild(guildId uint64) error {
	return rest.LeaveGuild(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) CreateDM(recipientId uint64) (channel.Channel, error) {
	return rest.CreateDM(context.Background(), ctx.Token, ctx.RateLimiter, recipientId)
}

func (ctx *Context) GetUserConnections() ([]integration.Connection, error) {
	return rest.GetUserConnections(context.Background(), ctx.Token, ctx.RateLimiter)
}

// GetGuildVoiceRegions should be preferred, as it returns VIP servers if available to the guild
func (ctx *Context) ListVoiceRegions() ([]guild.VoiceRegion, error) {
	return rest.ListVoiceRegions(context.Background(), ctx.Token)
}

func (ctx *Context) CreateWebhook(channelId uint64, data rest.WebhookData) (guild.Webhook, error) {
	return rest.CreateWebhook(context.Background(), ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) GetChannelWebhooks(channelId uint64) ([]guild.Webhook, error) {
	return rest.GetChannelWebhooks(context.Background(), ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) GetGuildWebhooks(guildId uint64) ([]guild.Webhook, error) {
	return rest.GetGuildWebhooks(context.Background(), ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetWebhook(webhookId uint64) (guild.Webhook, error) {
	return rest.GetWebhook(context.Background(), ctx.Token, ctx.RateLimiter, webhookId)
}

func (ctx *Context) ModifyWebhook(webhookId uint64, data rest.ModifyWebhookData) (guild.Webhook, error) {
	return rest.ModifyWebhook(context.Background(), ctx.Token, ctx.RateLimiter, webhookId, data)
}

func (ctx *Context) DeleteWebhook(webhookId uint64) error {
	return rest.DeleteWebhook(context.Background(), ctx.Token, ctx.RateLimiter, webhookId)
}

// if wait=true, a message object will be returned
func (ctx *Context) ExecuteWebhook(webhookId uint64, webhookToken string, wait bool, data rest.WebhookBody) (*message.Message, error) {
	return rest.ExecuteWebhook(context.Background(), webhookToken, ctx.RateLimiter, webhookId, wait, data)
}

func (ctx *Context) GetGuildAuditLog(guildId uint64, data rest.GetGuildAuditLogData) (auditlog.AuditLog, error) {
	return rest.GetGuildAuditLog(context.Background(), ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) GetGlobalCommands(applicationId uint64) ([]interaction.ApplicationCommand, error) {
	return rest.GetGlobalCommands(context.Background(), ctx.Token, ctx.RateLimiter, applicationId)
}

func (ctx *Context) CreateGlobalCommand(applicationId uint64, data rest.CreateCommandData) (interaction.ApplicationCommand, error) {
	return rest.CreateGlobalCommand(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, data)
}

func (ctx *Context) ModifyGlobalCommand(applicationId, commandId uint64, data rest.CreateCommandData) (interaction.ApplicationCommand, error) {
	return rest.ModifyGlobalCommand(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, commandId, data)
}

func (ctx *Context) DeleteGlobalCommand(applicationId, commandId uint64) error {
	return rest.DeleteGlobalCommand(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, commandId)
}

func (ctx *Context) GetGuildCommands(applicationId, guildId uint64) ([]interaction.ApplicationCommand, error) {
	return rest.GetGuildCommands(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId)
}

func (ctx *Context) CreateGuildCommand(applicationId, guildId uint64, data rest.CreateCommandData) (interaction.ApplicationCommand, error) {
	return rest.CreateGuildCommand(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId, data)
}

func (ctx *Context) ModifyGuildCommand(applicationId, guildId, commandId uint64, data rest.CreateCommandData) (interaction.ApplicationCommand, error) {
	return rest.ModifyGuildCommand(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId, commandId, data)
}

func (ctx *Context) DeleteGuildCommand(applicationId, guildId, commandId uint64) error {
	return rest.DeleteGuildCommand(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId, commandId)
}

func (ctx *Context) GetCommandPermissions(applicationId, guildId, commandId uint64) (rest.CommandWithPermissionsData, error) {
	return rest.GetCommandPermissions(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId, commandId)
}

func (ctx *Context) GetBulkCommandPermissions(applicationId, guildId uint64) ([]rest.CommandWithPermissionsData, error) {
	return rest.GetBulkCommandPermissions(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId)
}

func (ctx *Context) EditCommandPermissions(applicationId, guildId, commandId uint64, data rest.CommandWithPermissionsData) (rest.CommandWithPermissionsData, error) {
	return rest.EditCommandPermissions(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId, commandId, data)
}

func (ctx *Context) EditBulkCommandPermissions(applicationId, guildId uint64, data []rest.CommandWithPermissionsData) ([]rest.CommandWithPermissionsData, error) {
	return rest.EditBulkCommandPermissions(context.Background(), ctx.Token, ctx.RateLimiter, applicationId, guildId, data)
}
