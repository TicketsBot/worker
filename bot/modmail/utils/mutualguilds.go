package utils

import (
	"context"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/jackc/pgtype"
	"github.com/rxdn/gdl/objects/guild"
)

type UserGuild struct {
	Id   uint64
	Name string
}

func GetMutualGuilds(worker *worker.Context, userId uint64) []guild.Guild {
	var query string
	var args []interface{}
	if worker.IsWhitelabel {
		// get whitelabel guilds
		guilds, err := dbclient.Client.WhitelabelGuilds.GetGuilds(worker.BotId)
		if err != nil {
			sentry.Error(err)
			return nil
		}

		array := &pgtype.Int8Array{}
		if err := array.Set(guilds); err != nil {
			sentry.Error(err)
			return nil
		}

		query = `SELECT "guild_id" FROM members WHERE "user_id" = $1 AND "guild_id" = ANY($2);`
		args = []interface{}{userId, array}
	} else {
		query = `SELECT "guild_id" FROM members WHERE "user_id" = $1;`
		args = []interface{}{userId}
	}

	var guildIds []uint64
	rows, err := worker.Cache.Query(context.Background(), query, args...)
	defer rows.Close()
	if err != nil {
		sentry.Error(err)
		return nil
	}

	for rows.Next() {
		var guildId uint64
		if err := rows.Scan(&guildId); err != nil {
			sentry.Error(err)
			continue
		}

		guildIds = append(guildIds, guildId)
	}

	var guilds []guild.Guild
	for _, guildId := range guildIds {
		guild, err := worker.GetGuild(guildId); if err != nil {
			sentry.Error(err)
			continue
		}

		guilds = append(guilds, guild)
	}

	return guilds
}
