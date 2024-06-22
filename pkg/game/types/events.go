package types

import "github.com/cbodonnell/flywheel/pkg/kinematic"

type ConnectPlayerEvent struct {
	ClientID           uint32
	CharacterID        int32
	CharacterName      string
	CharacterPosition  kinematic.Vector
	CharacterFlipH     bool
	CharacterHitpoints int16
}

type DisconnectPlayerEvent struct {
	ClientID uint32
}
