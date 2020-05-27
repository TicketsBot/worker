package setup

import (
	"fmt"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
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

	msg := fmt.Sprintf("The setup has been completed!\n" +
		"You can add / remove support staff / admins using:\n" +
		"`t!addadmin [@User / Role Name]`\n" +
		"`t!removeadmin [@User / Role Name]`\n" +
		"`t!addsupport [@User / Role Name]`\n" +
		"`t!removesupprt [@User / Role Name]`\n" +
		"You can access more settings on the web panel at <https://panel.ticketsbot.net>\n" +
		"You should also consider creating a panel by visiting https://panel.ticketsbot.net/manage/%d/panels",
		u.Guild,
	)

	// Psuedo-premium
	utils.SendEmbed(u.Worker, u.Channel, utils.Green, "Setup", msg, nil, 30, true)
}

func (u *SetupUser) Cancel() {
	redis.Client.Del(fmt.Sprintf("setup:%s", u.ToString()))
}
