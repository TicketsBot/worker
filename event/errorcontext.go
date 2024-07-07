package event

import (
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
	"time"
)

type InteractionErrorContext struct {
	data map[string]string
}

func (ctx InteractionErrorContext) ToMap() map[string]string {
	return ctx.data
}

func NewApplicationCommandInteractionErrorContext(data interaction.ApplicationCommandInteraction) InteractionErrorContext {
	return InteractionErrorContext{
		data: map[string]string{
			"interaction_id":        strconv.FormatUint(data.Id, 10),
			"interaction_timestamp": utils.SnowflakeToTime(data.Id).String(),
			"current_time":          time.Now().String(),
			"command_name":          data.Data.Name,
		},
	}
}

func NewMessageComponentInteractionErrorContext(data interaction.InteractionMetadata) InteractionErrorContext {
	m := map[string]string{
		"interaction_id":        strconv.FormatUint(data.Id, 10),
		"interaction_timestamp": utils.SnowflakeToTime(data.Id).String(),
		"current_time":          time.Now().String(),
		//"component_type":        strconv.Itoa(int(data.Data.ComponentType)),
	}

	if !data.GuildId.IsNull {
		m["guild_id"] = strconv.FormatUint(data.GuildId.Value, 10)
	}

	if data.Member != nil {
		m["user_id"] = strconv.FormatUint(data.Member.User.Id, 10)
	} else if data.User != nil {
		m["user_id"] = strconv.FormatUint(data.User.Id, 10)
	}

	/*
		if data.Data.Type() == component.ComponentButton {
			m["custom_id"] = data.Data.AsButton().CustomId
		} else if data.Data.Type() == component.ComponentSelectMenu {
			m["custom_id"] = data.Data.AsSelectMenu().CustomId
		}
	*/

	return InteractionErrorContext{
		data: m,
	}
}
