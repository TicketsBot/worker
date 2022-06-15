package utils

import "fmt"

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

var (
	EmojiId         = NewCustomEmoji("id", 974006684643127296, false)
	EmojiOpen       = NewCustomEmoji("open", 974006684584378389, false)
	EmojiClose      = NewCustomEmoji("close", 974006684576002109, false)
	EmojiReason     = NewCustomEmoji("reason", 974006684567629845, false)
	EmojiTranscript = NewCustomEmoji("transcript", 974006684236267521, false)
	EmojiTime       = NewCustomEmoji("time", 974006684622159952, false)
	EmojiClaim      = NewCustomEmoji("claim", 974006684483715072, false)
)
