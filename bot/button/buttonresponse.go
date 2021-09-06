package button

import "github.com/TicketsBot/worker/bot/command"

type Response struct {
	Type ResponseType
	Data command.MessageResponse
}

type ResponseType uint8

const (
	ResponseTypeMessage ResponseType = iota
	ResponseTypeEdit
)
