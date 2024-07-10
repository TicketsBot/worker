package customisation

import (
	"context"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/dbclient"
)

type Colour int16

func (c Colour) Int16() int16 {
	return int16(c)
}

func (c Colour) Default() int {
	return DefaultColours[c]
}

const (
	Green Colour = iota
	Red
	Orange
	Lime
	Blue
)

var DefaultColours = map[Colour]int{
	Green:  0x2ECC71,
	Red:    0xFC3F35,
	Orange: 16740864,
	Lime:   7658240,
	Blue:   472219,
}

func GetDefaultColour(colour Colour) int {
	return DefaultColours[colour]
}

func IsValidColour(colour Colour) bool {
	_, valid := DefaultColours[colour]
	return valid
}

func GetColours(ctx context.Context, guildId uint64) (map[Colour]int, error) {
	raw, err := dbclient.Client.CustomColours.GetAll(ctx, guildId)
	if err != nil {
		return DefaultColours, err
	}

	colours := make(map[Colour]int)
	for id, hex := range raw {
		colours[Colour(id)] = hex
	}

	for id, hex := range DefaultColours {
		if _, ok := colours[id]; !ok {
			colours[id] = hex
		}
	}

	return colours, nil
}

// TODO: Premium check
func GetColour(ctx context.Context, guildId uint64, colourCode Colour) (int, error) {
	colour, ok, err := dbclient.Client.CustomColours.Get(ctx, guildId, colourCode.Int16())
	if err != nil {
		return 0, err
	}

	if !ok {
		return GetDefaultColour(colourCode), nil
	}

	return colour, nil
}

// TODO: Premium check
func GetColourOrDefault(ctx context.Context, guildId uint64, colourCode Colour) int {
	colour, ok, err := dbclient.Client.CustomColours.Get(ctx, guildId, colourCode.Int16())
	if err != nil {
		sentry.Error(err)
		return GetDefaultColour(colourCode)
	}

	if !ok {
		return GetDefaultColour(colourCode)
	}

	return colour
}
