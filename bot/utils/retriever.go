package utils

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
)

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
	return permission.NewRedisCache(redis.Client)
}

func (wr WorkerRetriever) IsBotAdmin(userId uint64) bool {
	return IsBotAdmin(userId)
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
