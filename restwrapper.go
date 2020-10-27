package worker

import (
	"github.com/rxdn/gdl/objects/auditlog"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/integration"
	"github.com/rxdn/gdl/objects/invite"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
)

func (ctx *Context) GetChannel(channelId uint64) (channel.Channel, error) {
	shouldCache := ctx.Cache.GetOptions().Channels
	if shouldCache {
		if cached, found := ctx.Cache.GetChannel(channelId); found {
			return cached, nil
		}
	}

	channel, err := rest.GetChannel(ctx.Token, ctx.RateLimiter, channelId)

	if shouldCache && err == nil {
		go ctx.Cache.StoreChannel(channel)
	}

	return channel, err
}

func (ctx *Context) ModifyChannel(channelId uint64, data rest.ModifyChannelData) (channel.Channel, error) {
	channel, err := rest.ModifyChannel(ctx.Token, ctx.RateLimiter, channelId, data)

	if ctx.Cache.GetOptions().Channels && err != nil {
		go ctx.Cache.StoreChannel(channel)
	}

	return channel, err
}

func (ctx *Context) DeleteChannel(channelId uint64) (channel.Channel, error) {
	return rest.DeleteChannel(ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) GetChannelMessages(channelId uint64, options rest.GetChannelMessagesData) ([]message.Message, error) {
	return rest.GetChannelMessages(ctx.Token, ctx.RateLimiter, channelId, options)
}

func (ctx *Context) GetChannelMessage(channelId, messageId uint64) (message.Message, error) {
	return rest.GetChannelMessage(ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) CreateMessage(channelId uint64, content string) (message.Message, error) {
	return ctx.CreateMessageComplex(channelId, rest.CreateMessageData{
		Content: content,
	})
}

func (ctx *Context) CreateMessageEmbed(channelId uint64, embed *embed.Embed) (message.Message, error) {
	return ctx.CreateMessageComplex(channelId, rest.CreateMessageData{
		Embed: embed,
	})
}

func (ctx *Context) CreateMessageComplex(channelId uint64, data rest.CreateMessageData) (message.Message, error) {
	return rest.CreateMessage(ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) CreateReaction(channelId, messageId uint64, emoji string) error {
	return rest.CreateReaction(ctx.Token, ctx.RateLimiter, channelId, messageId, emoji)
}

func (ctx *Context) DeleteOwnReaction(channelId, messageId uint64, emoji string) error {
	return rest.DeleteOwnReaction(ctx.Token, ctx.RateLimiter, channelId, messageId, emoji)
}

func (ctx *Context) DeleteUserReaction(channelId, messageId, userId uint64, emoji string) error {
	return rest.DeleteUserReaction(ctx.Token, ctx.RateLimiter, channelId, messageId, userId, emoji)
}

func (ctx *Context) GetReactions(channelId, messageId uint64, emoji string, options rest.GetReactionsData) ([]user.User, error) {
	return rest.GetReactions(ctx.Token, ctx.RateLimiter, channelId, messageId, emoji, options)
}

func (ctx *Context) DeleteAllReactions(channelId, messageId uint64) error {
	return rest.DeleteAllReactions(ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) DeleteAllReactionsEmoji(channelId, messageId uint64, emoji string) error {
	return rest.DeleteAllReactionsEmoji(ctx.Token, ctx.RateLimiter, channelId, messageId, emoji)
}

func (ctx *Context) EditMessage(channelId, messageId uint64, data rest.EditMessageData) (message.Message, error) {
	return rest.EditMessage(ctx.Token, ctx.RateLimiter, channelId, messageId, data)
}

func (ctx *Context) DeleteMessage(channelId, messageId uint64) error {
	return rest.DeleteMessage(ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) BulkDeleteMessages(channelId uint64, messages []uint64) error {
	return rest.BulkDeleteMessages(ctx.Token, ctx.RateLimiter, channelId, messages)
}

func (ctx *Context) EditChannelPermissions(channelId uint64, updated channel.PermissionOverwrite) error {
	return rest.EditChannelPermissions(ctx.Token, ctx.RateLimiter, channelId, updated)
}

func (ctx *Context) GetChannelInvites(channelId uint64) ([]invite.InviteMetadata, error) {
	return rest.GetChannelInvites(ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) CreateChannelInvite(channelId uint64, data rest.CreateInviteData) (invite.Invite, error) {
	return rest.CreateChannelInvite(ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) DeleteChannelPermissions(channelId, overwriteId uint64) error {
	return rest.DeleteChannelPermissions(ctx.Token, ctx.RateLimiter, channelId, overwriteId)
}

func (ctx *Context) TriggerTypingIndicator(channelId uint64) error {
	return rest.TriggerTypingIndicator(ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) GetPinnedMessages(channelId uint64) ([]message.Message, error) {
	return rest.GetPinnedMessages(ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) AddPinnedChannelMessage(channelId, messageId uint64) error {
	return rest.AddPinnedChannelMessage(ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) DeletePinnedChannelMessage(channelId, messageId uint64) error {
	return rest.DeletePinnedChannelMessage(ctx.Token, ctx.RateLimiter, channelId, messageId)
}

func (ctx *Context) ListGuildEmojis(guildId uint64) ([]emoji.Emoji, error) {
	shouldCacheEmoji := ctx.Cache.GetOptions().Emojis
	shouldCacheGuild := ctx.Cache.GetOptions().Guilds

	if shouldCacheEmoji && shouldCacheGuild {
		if guild, found := ctx.Cache.GetGuild(guildId, false); found {
			return guild.Emojis, nil
		}
	}

	emojis, err := rest.ListGuildEmojis(ctx.Token, ctx.RateLimiter, guildId)

	if shouldCacheEmoji && err == nil {
		go func() {
			for _, emoji := range emojis {
				ctx.Cache.StoreEmoji(emoji, guildId)
			}
		}()
	}

	return emojis, err
}

func (ctx *Context) GetGuildEmoji(guildId uint64, emojiId uint64) (emoji.Emoji, error) {
	shouldCache := ctx.Cache.GetOptions().Emojis
	if shouldCache {
		if emoji, found := ctx.Cache.GetEmoji(emojiId); found {
			return emoji, nil
		}
	}

	emoji, err := rest.GetGuildEmoji(ctx.Token, ctx.RateLimiter, guildId, emojiId)

	if shouldCache && err == nil {
		go ctx.Cache.StoreEmoji(emoji, guildId)
	}

	return emoji, err
}

func (ctx *Context) CreateGuildEmoji(guildId uint64, data rest.CreateEmojiData) (emoji.Emoji, error) {
	return rest.CreateGuildEmoji(ctx.Token, ctx.RateLimiter, guildId, data)
}

// updating Image is not permitted
func (ctx *Context) ModifyGuildEmoji(guildId, emojiId uint64, data rest.CreateEmojiData) (emoji.Emoji, error) {
	return rest.ModifyGuildEmoji(ctx.Token, ctx.RateLimiter, guildId, emojiId, data)
}

func (ctx *Context) CreateGuild(data rest.CreateGuildData) (guild.Guild, error) {
	return rest.CreateGuild(ctx.Token, data)
}

func (ctx *Context) GetGuild(guildId uint64) (guild.Guild, error) {
	shouldCache := ctx.Cache.GetOptions().Guilds

	if shouldCache {
		if cachedGuild, found := ctx.Cache.GetGuild(guildId, false); found {
			return cachedGuild, nil
		}
	}

	guild, err := rest.GetGuild(ctx.Token, ctx.RateLimiter, guildId)
	if err == nil {
		go ctx.Cache.StoreGuild(guild)
	}

	return guild, err
}

func (ctx *Context) GetGuildPreview(guildId uint64) (guild.GuildPreview, error) {
	return rest.GetGuildPreview(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) ModifyGuild(guildId uint64, data rest.ModifyGuildData) (guild.Guild, error) {
	return rest.ModifyGuild(ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) DeleteGuild(guildId uint64) error {
	return rest.DeleteGuild(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetGuildChannels(guildId uint64) ([]channel.Channel, error) {
	shouldCache := ctx.Cache.GetOptions().Guilds && ctx.Cache.GetOptions().Channels

	if shouldCache {
		cached := ctx.Cache.GetGuildChannels(guildId)

		// either not cached (more likely), or guild has no channels
		if len(cached) > 0 {
			return cached, nil
		}
	}

	channels, err := rest.GetGuildChannels(ctx.Token, ctx.RateLimiter, guildId)

	if shouldCache && err == nil {
		go func() {
			for _, channel := range channels {
				ctx.Cache.StoreChannel(channel)
			}
		}()
	}

	return channels, err
}

func (ctx *Context) CreateGuildChannel(guildId uint64, data rest.CreateChannelData) (channel.Channel, error) {
	return rest.CreateGuildChannel(ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) ModifyGuildChannelPositions(guildId uint64, positions []rest.Position) error {
	return rest.ModifyGuildChannelPositions(ctx.Token, ctx.RateLimiter, guildId, positions)
}

func (ctx *Context) GetGuildMember(guildId, userId uint64) (member.Member, error) {
	cacheGuilds := ctx.Cache.GetOptions().Guilds
	cacheUsers := ctx.Cache.GetOptions().Users

	if cacheGuilds && cacheUsers {
		if member, found := ctx.Cache.GetMember(guildId, userId); found {
			return member, nil
		}
	}

	member, err := rest.GetGuildMember(ctx.Token, ctx.RateLimiter, guildId, userId)

	if cacheGuilds && err == nil {
		go ctx.Cache.StoreMember(member, guildId)
	}

	return member, err
}

func (ctx *Context) ListGuildMembers(guildId uint64, data rest.ListGuildMembersData) ([]member.Member, error) {
	members, err := rest.ListGuildMembers(ctx.Token, ctx.RateLimiter, guildId, data)
	if err == nil {
		go func() {
			for _, member := range members {
				ctx.Cache.StoreMember(member, guildId)
			}
		}()
	}

	return members, err
}

func (ctx *Context) ModifyGuildMember(guildId, userId uint64, data rest.ModifyGuildMemberData) error {
	return rest.ModifyGuildMember(ctx.Token, ctx.RateLimiter, guildId, userId, data)
}

func (ctx *Context) ModifyCurrentUserNick(guildId uint64, nick string) error {
	return rest.ModifyCurrentUserNick(ctx.Token, ctx.RateLimiter, guildId, nick)
}

func (ctx *Context) AddGuildMemberRole(guildId, userId, roleId uint64) error {
	return rest.AddGuildMemberRole(ctx.Token, ctx.RateLimiter, guildId, userId, roleId)
}

func (ctx *Context) RemoveGuildMemberRole(guildId, userId, roleId uint64) error {
	return rest.RemoveGuildMemberRole(ctx.Token, ctx.RateLimiter, guildId, userId, roleId)
}

func (ctx *Context) RemoveGuildMember(guildId, userId uint64) error {
	return rest.RemoveGuildMember(ctx.Token, ctx.RateLimiter, guildId, userId)
}

func (ctx *Context) GetGuildBans(guildId uint64) ([]guild.Ban, error) {
	return rest.GetGuildBans(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetGuildBan(guildId, userId uint64) (guild.Ban, error) {
	return rest.GetGuildBan(ctx.Token, ctx.RateLimiter, guildId, userId)
}

func (ctx *Context) CreateGuildBan(guildId, userId uint64, data rest.CreateGuildBanData) error {
	return rest.CreateGuildBan(ctx.Token, ctx.RateLimiter, guildId, userId, data)
}

func (ctx *Context) RemoveGuildBan(guildId, userId uint64) error {
	return rest.RemoveGuildBan(ctx.Token, ctx.RateLimiter, guildId, userId)
}

func (ctx *Context) GetGuildRoles(guildId uint64) ([]guild.Role, error) {
	shouldCache := ctx.Cache.GetOptions().Guilds
	if shouldCache {
		cached := ctx.Cache.GetGuildRoles(guildId)

		// either not cached (more likely), or guild has no channels
		if len(cached) > 0 {
			return cached, nil
		}
	}

	roles, err := rest.GetGuildRoles(ctx.Token, ctx.RateLimiter, guildId)

	if shouldCache && err == nil {
		go func() {
			for _, role := range roles {
				ctx.Cache.StoreRole(role, guildId)
			}
		}()
	}

	return roles, err
}

func (ctx *Context) CreateGuildRole(guildId uint64, data rest.GuildRoleData) (guild.Role, error) {
	return rest.CreateGuildRole(ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) ModifyGuildRolePositions(guildId uint64, positions []rest.Position) ([]guild.Role, error) {
	return rest.ModifyGuildRolePositions(ctx.Token, ctx.RateLimiter, guildId, positions)
}

func (ctx *Context) ModifyGuildRole(guildId, roleId uint64, data rest.GuildRoleData) (guild.Role, error) {
	return rest.ModifyGuildRole(ctx.Token, ctx.RateLimiter, guildId, roleId, data)
}

func (ctx *Context) DeleteGuildRole(guildId, roleId uint64) error {
	return rest.DeleteGuildRole(ctx.Token, ctx.RateLimiter, guildId, roleId)
}

func (ctx *Context) GetGuildPruneCount(guildId uint64, days int) (int, error) {
	return rest.GetGuildPruneCount(ctx.Token, ctx.RateLimiter, guildId, days)
}

// computePruneCount = whether 'pruned' is returned, discouraged for large guilds
func (ctx *Context) BeginGuildPrune(guildId uint64, days int, computePruneCount bool) error {
	return rest.BeginGuildPrune(ctx.Token, ctx.RateLimiter, guildId, days, computePruneCount)
}

func (ctx *Context) GetGuildVoiceRegions(guildId uint64) ([]guild.VoiceRegion, error) {
	return rest.GetGuildVoiceRegions(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetGuildInvites(guildId uint64) ([]invite.InviteMetadata, error) {
	return rest.GetGuildInvites(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetGuildIntegrations(guildId uint64) ([]integration.Integration, error) {
	return rest.GetGuildIntegrations(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) CreateGuildIntegration(guildId uint64, data rest.CreateIntegrationData) error {
	return rest.CreateGuildIntegration(ctx.Token, ctx.RateLimiter, guildId, data)
}

func (ctx *Context) ModifyGuildIntegration(guildId, integrationId uint64, data rest.ModifyIntegrationData) error {
	return rest.ModifyGuildIntegration(ctx.Token, ctx.RateLimiter, guildId, integrationId, data)
}

func (ctx *Context) DeleteGuildIntegration(guildId, integrationId uint64) error {
	return rest.DeleteGuildIntegration(ctx.Token, ctx.RateLimiter, guildId, integrationId)
}

func (ctx *Context) SyncGuildIntegration(guildId, integrationId uint64) error {
	return rest.SyncGuildIntegration(ctx.Token, ctx.RateLimiter, guildId, integrationId)
}

func (ctx *Context) GetGuildEmbed(guildId uint64) (guild.GuildWidget, error) {
	return rest.GetGuildWidget(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) ModifyGuildEmbed(guildId uint64, data guild.GuildEmbed) (guild.GuildEmbed, error) {
	return rest.ModifyGuildEmbed(ctx.Token, ctx.RateLimiter, guildId, data)
}

// returns invite object with only "code" and "uses" fields
func (ctx *Context) GetGuildVanityUrl(guildId uint64) (invite.Invite, error) {
	return rest.GetGuildVanityURL(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetInvite(inviteCode string, withCounts bool) (invite.Invite, error) {
	return rest.GetInvite(ctx.Token, ctx.RateLimiter, inviteCode, withCounts)
}

func (ctx *Context) DeleteInvite(inviteCode string) (invite.Invite, error) {
	return rest.DeleteInvite(ctx.Token, ctx.RateLimiter, inviteCode)
}

func (ctx *Context) GetCurrentUser() (user.User, error) {
	if cached, found := ctx.Cache.GetSelf(); found {
		return cached, nil
	}

	self, err := rest.GetCurrentUser(ctx.Token, ctx.RateLimiter)

	if err == nil {
		go ctx.Cache.StoreSelf(self)
	}

	return self, err
}

func (ctx *Context) GetUser(userId uint64) (user.User, error) {
	shouldCache := ctx.Cache.GetOptions().Users

	if shouldCache {
		if cached, found := ctx.Cache.GetUser(userId); found {
			return cached, nil
		}
	}

	user, err := rest.GetUser(ctx.Token, ctx.RateLimiter, userId)

	if shouldCache && err == nil {
		go ctx.Cache.StoreUser(user)
	}

	return user, err
}

func (ctx *Context) ModifyCurrentUser(data rest.ModifyUserData) (user.User, error) {
	return rest.ModifyCurrentUser(ctx.Token, ctx.RateLimiter, data)
}

func (ctx *Context) GetCurrentUserGuilds(data rest.CurrentUserGuildsData) ([]guild.Guild, error) {
	return rest.GetCurrentUserGuilds(ctx.Token, ctx.RateLimiter, data)
}

func (ctx *Context) LeaveGuild(guildId uint64) error {
	return rest.LeaveGuild(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) CreateDM(recipientId uint64) (channel.Channel, error) {
	return rest.CreateDM(ctx.Token, ctx.RateLimiter, recipientId)
}

func (ctx *Context) GetUserConnections() ([]integration.Connection, error) {
	return rest.GetUserConnections(ctx.Token, ctx.RateLimiter)
}

// GetGuildVoiceRegions should be preferred, as it returns VIP servers if available to the guild
func (ctx *Context) ListVoiceRegions() ([]guild.VoiceRegion, error) {
	return rest.ListVoiceRegions(ctx.Token)
}

func (ctx *Context) CreateWebhook(channelId uint64, data rest.WebhookData) (guild.Webhook, error) {
	return rest.CreateWebhook(ctx.Token, ctx.RateLimiter, channelId, data)
}

func (ctx *Context) GetChannelWebhooks(channelId uint64) ([]guild.Webhook, error) {
	return rest.GetChannelWebhooks(ctx.Token, ctx.RateLimiter, channelId)
}

func (ctx *Context) GetGuildWebhooks(guildId uint64) ([]guild.Webhook, error) {
	return rest.GetGuildWebhooks(ctx.Token, ctx.RateLimiter, guildId)
}

func (ctx *Context) GetWebhook(webhookId uint64) (guild.Webhook, error) {
	return rest.GetWebhook(ctx.Token, ctx.RateLimiter, webhookId)
}

func (ctx *Context) ModifyWebhook(webhookId uint64, data rest.ModifyWebhookData) (guild.Webhook, error) {
	return rest.ModifyWebhook(ctx.Token, ctx.RateLimiter, webhookId, data)
}

func (ctx *Context) DeleteWebhook(webhookId uint64) error {
	return rest.DeleteWebhook(ctx.Token, ctx.RateLimiter, webhookId)
}

// if wait=true, a message object will be returned
func (ctx *Context) ExecuteWebhook(webhookId uint64, webhookToken string, wait bool, data rest.WebhookBody) (*message.Message, error) {
	return rest.ExecuteWebhook(webhookToken, ctx.RateLimiter, webhookId, wait, data)
}

func (ctx *Context) GetGuildAuditLog(guildId uint64, data rest.GetGuildAuditLogData) (auditlog.AuditLog, error) {
	return rest.GetGuildAuditLog(ctx.Token, ctx.RateLimiter, guildId, data)
}
