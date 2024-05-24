package messages

import (
	"encoding/json"

	"github.com/cbodonnell/flywheel/pkg/kinematic"
)

const (
	// UDPMessageBufferSize represents the maximum size of a message
	UDPMessageBufferSize = 508

	// TCPMessageBufferSize represents the maximum size of a message
	TCPMessageBufferSize = 1460
)

// Message types
type MessageType uint8

const (
	MessageTypeClientLogin MessageType = iota
	MessageTypeServerLoginSuccess
	MessageTypeServerLoginFailure
	MessageTypeClientPing
	MessageTypeServerPong
	MessageTypeClientPlayerUpdate
	MessageTypeServerGameUpdate
	MessageTypeClientSyncTime
	MessageTypeServerSyncTime
	MessageTypeServerPlayerConnect
	MessageTypeServerPlayerDisconnect
	MessageTypeServerNPCHit
	MessageTypeServerNPCKill
	MessageTypeServerPlayerHit
	MessageTypeServerPlayerKill
)

func (m MessageType) String() string {
	return [...]string{
		"ClientLogin",
		"ServerLoginSuccess",
		"ServerLoginFailure",
		"ClientPing",
		"ServerPong",
		"ClientPlayerUpdate",
		"ServerGameUpdate",
		"ClientSyncTime",
		"ServerSyncTime",
		"ServerPlayerConnect",
		"ServerPlayerDisconnect",
		"ServerNPCHit",
		"ServerNPCKill",
		"ServerPlayerHit",
		"ServerPlayerKill",
	}[m]
}

// Message represents a generic message for serialization/deserialization
type Message struct {
	ClientID uint32          `json:"clientID"`
	Type     MessageType     `json:"type"`
	Payload  json.RawMessage `json:"payload"`
}

type ClientLogin struct {
	Token       string `json:"token"`
	CharacterID int32  `json:"characterID"`
}

type ServerLoginSuccess struct {
	ClientID uint32 `json:"clientID"`
}

type ServerLoginFailure struct {
	Reason string `json:"reason"`
}

const (
	// MaxPreviousUpdates is the maximum number of previous updates to send to the server
	// TODO: determine a more appropriate size
	MaxPreviousUpdates = 2
)

// TODO: split this into PlayerInput and ClientPlayerUpdate to differentiate the game and message types
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
	// InputAttack1 is the attack 1 input from the client
	InputAttack1 bool `json:"inputAttack1"`
	// InputAttack2 is the attack 2 input from the client
	InputAttack2 bool `json:"inputAttack2"`
	// InputAttack3 is the attack 3 input from the client
	InputAttack3 bool `json:"inputAttack3"`
	// InputRespawn is the respawn input from the client
	InputRespawn bool `json:"inputRespawn"`
	// DeltaTime is the time since the last update as recorded by the client
	DeltaTime float64 `json:"deltaTime"`
	// PastUpdates is a list of past updates from the client to
	// mitigate the effects of packet loss and out-of-order delivery
	PastUpdates []*ClientPlayerUpdate `json:"previousUpdates"`
}

// ServerGameUpdate is a message sent by the server to update clients on the game state
type ServerGameUpdate struct {
	// Timestamp is the time at which the update was generated by the server
	Timestamp int64 `json:"timestamp"`
	// Players maps client IDs to player states
	Players map[uint32]*PlayerStateUpdate `json:"players"`
	// NPCs maps enemy IDs to NPC states
	NPCs map[uint32]*NPCStateUpdate `json:"npcs"`
}

// PlayerStateUpdate is a message sent by the server to update clients on a player's state
type PlayerStateUpdate struct {
	// LastProcessedTimestamp is the timestamp of the last processed update
	LastProcessedTimestamp int64 `json:"lastProcessedTimestamp"`
	// CharacterID is the ID of the player's character
	CharacterID int32 `json:"characterID"`
	// Name is the name of the player
	Name string `json:"playerName"`
	// Position is the position of the player
	Position kinematic.Vector `json:"position"`
	// Velocity is the velocity of the player
	Velocity kinematic.Vector `json:"velocity"`
	// IsOnGround is a flag indicating whether the player is on the ground
	IsOnGround bool `json:"isOnGround"`
	// IsAttacking is a flag indicating whether the player is attacking
	IsAttacking bool `json:"isAttacking"`
	// Animation is the current animation of the player
	Animation uint8 `json:"animation"`
	// AnimationFlip is a flag indicating whether the animation should be flipped
	AnimationFlip bool `json:"animationFlip"`
	// AnimationSequence is used to differentiate between different instances of the same animation
	AnimationSequence uint8 `json:"animationSequence"`
	// Hitpoints is the current hitpoints of the player
	Hitpoints int16 `json:"hitpoints"`
}

// NPCStateUpdate is a message sent by the server to update clients on an NPC's state
type NPCStateUpdate struct {
	// Position is the position of the NPC
	Position kinematic.Vector `json:"position"`
	// Velocity is the velocity of the NPC
	Velocity kinematic.Vector `json:"velocity"`
	// IsOnGround is a flag indicating whether the NPC is on the ground
	IsOnGround bool `json:"isOnGround"`
	// Animation is the current animation of the NPC
	Animation uint8 `json:"animation"`
	// AnimationFlip is a flag indicating whether the animation should be flipped
	AnimationFlip bool `json:"animationFlip"`
	// AnimationSequence is used to differentiate between different instances of the same animation
	AnimationSequence uint8 `json:"animationSequence"`
	// Hitpoints is the current hitpoints of the NPC
	Hitpoints int16 `json:"hitpoints"`
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
	PlayerState *PlayerStateUpdate `json:"playerState"`
}

// ServerPlayerDisconnect is a message sent by the server to notify clients that a player has disconnected
type ServerPlayerDisconnect struct {
	// ClientID is the ID of the player that has disconnected
	ClientID uint32 `json:"clientID"`
}

// ServerNPCHit is a message sent by the server to notify clients that an NPC has been hit
type ServerNPCHit struct {
	// NPCID is the ID of the NPC that has been hit
	NPCID uint32 `json:"npcID"`
	// PlayerID is the ID of the player that hit the NPC
	PlayerID uint32 `json:"playerID"`
	// Damage is the amount of damage dealt to the NPC
	Damage int16 `json:"damage"`
}

// ServerNPCKill is a message sent by the server to notify clients that an NPC has been killed
type ServerNPCKill struct {
	// NPCID is the ID of the NPC that has been killed
	NPCID uint32 `json:"npcID"`
	// PlayerID is the ID of the player that killed the NPC
	PlayerID uint32 `json:"playerID"`
}

// ServerPlayerHit is a message sent by the server to notify clients that a player has been hit
type ServerPlayerHit struct {
	// PlayerID is the ID of the player that has been hit
	PlayerID uint32 `json:"playerID"`
	// NPCID is the ID of the NPC that hit the player
	NPCID uint32 `json:"npcID"`
	// Damage is the amount of damage dealt to the player
	Damage int16 `json:"damage"`
}

// ServerPlayerKill is a message sent by the server to notify clients that a player has been killed
type ServerPlayerKill struct {
	// PlayerID is the ID of the player that has been killed
	PlayerID uint32 `json:"playerID"`
	// NPCID is the ID of the NPC that killed the player
	NPCID uint32 `json:"npcID"`
}
