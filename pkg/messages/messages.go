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
	MessageTypeServerAssignID         = "aid"
	MessageTypeClientPing             = "cp"
	MessageTypeServerPong             = "sp"
	MessageTypeClientPlayerUpdate     = "cpu"
	MessageTypeServerGameUpdate       = "sgu"
	MessageTypeClientSyncTime         = "cst"
	MessageTypeServerSyncTime         = "sst"
	MessageTypeServerPlayerConnect    = "spc"
	MessageTypeServerPlayerDisconnect = "spd"
)

// Message represents a generic message for serialization/deserialization
type Message struct {
	ClientID uint32          `json:"clientID"`
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"`
}

type AssignID struct {
	ClientID uint32 `json:"clientID"`
}

// TODO: change this to a more generic client input message
// with a type field to differentiate between different input types (move, jump, fire, etc.)
type ClientPlayerUpdate struct {
	// Timestamp is the time at which the update was generated by the client
	Timestamp int64 `json:"timestamp"`
	// InputX is the x-axis input from the client ranging from -1 to 1
	InputX float64 `json:"inputX"`
	// InputY is the y-axis input from the client ranging from -1 to 1
	InputY float64 `json:"inputY"`
	// InputJump is the jump input from the client
	InputJump bool `json:"inputJump"`
	// DeltaTime is the time since the last update as recorded by the client
	DeltaTime float64 `json:"deltaTime"`
}

// ClientSyncTime is a message sent by the client to request a time sync with the server
type ClientSyncTime struct {
	// Timestamp is the time at which the sync request was generated by the client
	Timestamp int64 `json:"timestamp"`
}

// ServerSyncTime is a message sent by the server in response to a time sync request
type ServerSyncTime struct {
	// Timestamp is the time at which the sync response was generated by the server
	Timestamp int64 `json:"timestamp"`
	// ClientTimestamp is the time at which the sync request was generated by the client
	ClientTimestamp int64 `json:"clientTimestamp"`
}

// ServerPlayerConnect is a message sent by the server to notify clients that a new player has connected
type ServerPlayerConnect struct {
	// ClientID is the ID of the player that has connected
	ClientID uint32 `json:"clientID"`
	// PlayerState is the state of the player that has connected
	PlayerState *gametypes.PlayerState `json:"playerState"`
}

// ServerPlayerDisconnect is a message sent by the server to notify clients that a player has disconnected
type ServerPlayerDisconnect struct {
	// ClientID is the ID of the player that has disconnected
	ClientID uint32 `json:"clientID"`
}
