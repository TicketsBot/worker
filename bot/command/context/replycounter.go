package context

import (
	"errors"
	"sync/atomic"
)

const InteractionReplyLimit = 5

type ReplyCounter struct {
	count atomic.Int32
}

var ErrReplyLimitReached = errors.New("reply limit reached")

func NewReplyCounter() *ReplyCounter {
	return &ReplyCounter{}
}

func (r *ReplyCounter) Increment() int {
	return int(r.count.Add(1))
}

func (r *ReplyCounter) Try() error {
	if r.count.Add(1) > InteractionReplyLimit {
		return ErrReplyLimitReached
	}

	return nil
}
