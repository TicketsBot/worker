package setup

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strconv"
	"time"
)

const timeout = time.Minute * 2

func (u *SetupUser) InSetup() bool {
	key := fmt.Sprintf("setup:%s", u.ToString())
	res, err := redis.Client.Exists(key).Result()
	if err != nil {
		return false
	}

	return res == 1
}

func (u *SetupUser) GetState() *State {
	key := fmt.Sprintf("setup:%s", u.ToString())
	res, err := redis.Client.Get(key).Result(); if err != nil {
		return nil
	}

	id, err := strconv.Atoi(res); if err != nil {
		return nil
	}

	state := State(id)
	return &state
}

func (u *SetupUser) Next() {
	state := u.GetState()

	var newState State
	if state == nil {
		newState = State(0)
	} else {
		id := int(*state) + 1

		if id > GetMaxStage() {
			u.Finish()
			return
		}

		newState = State(id)
	}

	key := fmt.Sprintf("setup:%s", u.ToString())
	redis.Client.Set(key, int(newState), timeout)
}

func (u *SetupUser) Finish() {
	redis.Client.Del(fmt.Sprintf("setup:%s", u.ToString()))

	embed := embed.NewEmbed().
		SetTitle("Setup Complete").
		SetColor(int(utils.Green)).
		SetDescription("The setup process has been completed; however you may like to look into creating a reaction panel or adding staff:").
		AddField("Reaction Panels", fmt.Sprintf("Reaction panels are a commonly used feature of the bot. You can read about them [here](https://ticketsbot.net/panels), or create one on [the dashboard](https://panel.ticketsbot.net/manage/%d/panels)", u.Guild), false).
		AddField("Adding Staff", "To make staff able to answer tickets, you must let the bot know about them first. You can do this through\n`t!addsupport [@User / @Role]` and `t!addadmin [@User / @Role]`. Administrators can change the settings of the bot and access the dashboard.", false)

	if _, err := u.Worker.CreateMessageEmbed(u.Channel, embed); err != nil {
		sentry.Error(err)
		return
	}
}

func (u *SetupUser) Cancel() {
	redis.Client.Del(fmt.Sprintf("setup:%s", u.ToString()))
}
