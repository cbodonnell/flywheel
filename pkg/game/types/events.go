package types

type ConnectPlayerEvent struct {
	ClientID    uint32
	PlayerState *PlayerState
}

type DisconnectPlayerEvent struct {
	ClientID uint32
}
