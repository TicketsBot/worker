package customisation

import (
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
)

func GetColourForGuild(worker *worker.Context, colour Colour, guildId uint64) (int, error) {
	premiumTier, err := utils.PremiumClient.GetTierByGuildId(guildId, true, worker.Token, worker.RateLimiter)
	if err != nil {
		return 0, err
	}

	if premiumTier > premium.None {
		colourCode, ok, err := dbclient.Client.CustomColours.Get(guildId, colour.Int16())
		if err != nil {
			return 0, err
		} else if !ok {
			return DefaultColours[colour], nil
		} else {
			return colourCode, nil
		}
	} else {
		return colour.Default(), nil
	}
}
