package utils

import (
	"fmt"
	"github.com/rxdn/gdl/objects"
	"github.com/rxdn/gdl/objects/guild/emoji"
)

type CustomEmoji struct {
	Name     string
	Id       uint64
	Animated bool
}

func NewCustomEmoji(name string, id uint64, animated bool) CustomEmoji {
	return CustomEmoji{
		Name: name,
		Id:   id,
	}
}

func (e CustomEmoji) String() string {
	if e.Animated {
		return fmt.Sprintf("<a:%s:%d>", e.Name, e.Id)
	} else {
		return fmt.Sprintf("<:%s:%d>", e.Name, e.Id)
	}
}

func (e CustomEmoji) BuildEmoji() *emoji.Emoji {
	return &emoji.Emoji{
		Id:       objects.NewNullableSnowflake(e.Id),
		Name:     e.Name,
		Animated: e.Animated,
	}
}

var (
	EmojiId         = NewCustomEmoji("id", 1013527224722391181, false)
	EmojiOpen       = NewCustomEmoji("open", 1013527364455649430, false)
	EmojiOpenTime   = NewCustomEmoji("opentime", 1013527365638430790, false)
	EmojiClose      = NewCustomEmoji("close", 1013527306192560188, false)
	EmojiCloseTime  = NewCustomEmoji("closetime", 1013527317341012009, false)
	EmojiReason     = NewCustomEmoji("reason", 1013527372399657023, false)
	EmojiSubject    = NewCustomEmoji("subject", 1013527369832738907, false)
	EmojiTranscript = NewCustomEmoji("transcript", 1013527375327281213, false)
	EmojiClaim      = NewCustomEmoji("claim", 1013527266124369980, false)
	EmojiPanel      = NewCustomEmoji("panel", 1013527367265820682, false)
	EmojiRating     = NewCustomEmoji("rating", 1013527368360538244, false)
	EmojiStaff      = NewCustomEmoji("staff", 1013527371216867370, false)
	EmojiThread     = NewCustomEmoji("thread", 1013527373750214717, false)
	EmojiBulletLine = NewCustomEmoji("bulletline", 1014161470491201596, false)
	//EmojiTime       = NewCustomEmoji("time", 974006684622159952, false)
)

// PrefixWithEmoji Useful for whitelabel bots
func PrefixWithEmoji(s string, emoji CustomEmoji, includeEmoji bool) string {
	if includeEmoji {
		return fmt.Sprintf("%s %s", emoji, s)
	} else {
		return s
	}
}
