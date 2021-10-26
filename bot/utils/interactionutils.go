package utils

import (
	"fmt"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

func ButtonInteractionUser(data interaction.MessageComponentInteraction) uint64 {
	if data.User != nil {
		return data.User.Id
	} else if data.Member != nil {
		return data.Member.User.Id
	} else { // Impossible
		return 0
	}
}

func StringChoice(value string) interaction.ApplicationCommandOptionChoice {
	return interaction.ApplicationCommandOptionChoice{
		Name:  value,
		Value: value,
	}
}

func IntChoice(value int) interaction.ApplicationCommandOptionChoice {
	return interaction.ApplicationCommandOptionChoice{
		Name:  strconv.Itoa(value),
		Value: value,
	}
}

func FloatChoice(value float32) interaction.ApplicationCommandOptionChoice {
	return interaction.ApplicationCommandOptionChoice{
		Name:  fmt.Sprintf("%f", value),
		Value: value,
	}
}
