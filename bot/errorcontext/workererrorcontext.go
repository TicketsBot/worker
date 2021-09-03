package errorcontext

import (
	"strconv"
)

type WorkerErrorContext struct {
	Guild       uint64
	User        uint64
	Channel     uint64
}

func (w WorkerErrorContext) ToMap() map[string]string {
	m := make(map[string]string)

	if w.Guild != 0 {
		m["guild"] = strconv.FormatUint(w.Guild, 10)
	}

	if w.User != 0 {
		m["user"] = strconv.FormatUint(w.User, 10)
	}

	if w.Channel != 0 {
		m["channel"] = strconv.FormatUint(w.Channel, 10)
	}

	return m
}
