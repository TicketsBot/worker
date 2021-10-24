package utils

import "github.com/rxdn/gdl/objects/interaction"

func ButtonInteractionUser(data interaction.MessageComponentInteraction) uint64 {
	if data.User != nil {
		return data.User.Id
	} else if data.Member != nil {
		return data.Member.User.Id
	} else { // Impossible
		return 0
	}
}
