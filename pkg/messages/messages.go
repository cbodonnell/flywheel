package messages

import (
	"encoding/json"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

const (
	// MessageBufferSize represents the maximum size of a message
	MessageBufferSize = 1024
)

// Message types
const (
	MessageTypeServerAssignID     = "aid"
	MessageTypeClientPing         = "cp"
	MessageTypeServerPong         = "sp"
	MessageTypeClientPlayerUpdate = "cpu"
	MessageTypeServerGameUpdate   = "sgu"
)

// Message represents a generic message for serialization/deserialization
type Message struct {
	ClientID uint32          `json:"clientID"`
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"`
}

type ClientPlayerUpdate struct {
	// Timestamp is the client time at which position is recorded
	Timestamp   int64                  `json:"timestamp"`
	PlayerState *gametypes.PlayerState `json:"playerState"`
}
