package redis

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const TicketStatusCacheExpiry = time.Second * 90

var ErrTicketStatusNotCached = errors.New("ticket status not cached")

func IsTicketChannel(ctx context.Context, channelId uint64) (bool, error) {
	key := fmt.Sprintf("isticket:%d", channelId)
	res, err := Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, ErrNil) {
			return false, ErrTicketStatusNotCached
		}

		return false, err
	}

	return res == "1", nil
}

func SetTicketChannelStatus(ctx context.Context, channelId uint64, isTicket bool) error {
	key := fmt.Sprintf("isticket:%d", channelId)

	var value string
	if isTicket {
		value = "1"
	} else {
		value = "0"
	}

	return Client.Set(ctx, key, value, TicketStatusCacheExpiry).Err()
}
