package utils

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
)

var cache = permission.NewMemoryCache()

func ToRetriever(worker *worker.Context) permission.Retriever {
	return WorkerRetriever{
		ctx: worker,
	}
}

type WorkerRetriever struct {
	ctx *worker.Context
}

func (wr WorkerRetriever) Db() *database.Database {
	return dbclient.Client
}

func (wr WorkerRetriever) Cache() permission.PermissionCache {
	return cache
}

func (wr WorkerRetriever) IsBotAdmin(userId uint64) bool {
	return IsBotAdmin(userId)
}

func (wr WorkerRetriever) GetGuild(guildId uint64) (guild.Guild, error) {
	return wr.ctx.GetGuild(guildId)
}

func (wr WorkerRetriever) GetChannel(channelId uint64) (channel.Channel, error) {
	return wr.ctx.GetChannel(channelId)
}

func (wr WorkerRetriever) GetGuildMember(guildId, userId uint64) (member.Member, error) {
	return wr.ctx.GetGuildMember(guildId, userId)
}

func (wr WorkerRetriever) GetGuildRoles(guildId uint64) ([]guild.Role, error) {
	return wr.ctx.GetGuildRoles(guildId)
}

func (wr WorkerRetriever) GetGuildOwner(guildId uint64) (uint64, error) {
	cachedOwner, exists := wr.ctx.Cache.GetGuildOwner(guildId)
	if exists {
		return cachedOwner, nil
	}

	guild, err := wr.ctx.GetGuild(guildId)
	if err != nil {
		return 0, err
	}

	go wr.ctx.Cache.StoreGuild(guild)
	return guild.OwnerId, nil
}
