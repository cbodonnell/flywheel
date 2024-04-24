package types

import "github.com/cbodonnell/flywheel/pkg/kinematic"

type ConnectPlayerEvent struct {
	ClientID uint32
	PlayerID string
	Position *kinematic.Vector
}

type DisconnectPlayerEvent struct {
	ClientID uint32
}
