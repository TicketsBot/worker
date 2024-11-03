package rpc

type Listener interface {
	HandleMessage(message []byte)
}

var Listeners = map[string]Listener{}
