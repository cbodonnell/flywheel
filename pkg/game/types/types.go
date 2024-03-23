package types

import "github.com/solarlune/resolv"

type GameState struct {
	// Timestamp is the time at which the game state was generated
	Timestamp int64 `json:"timestamp"`
	// Players maps client IDs to player states
	Players map[uint32]*PlayerState `json:"players"`
}

type PlayerState struct {
	LastProcessedTimestamp int64    `json:"lastProcessedTimestamp"`
	Position               Position `json:"position"`
	Velocity               Velocity `json:"velocity"`
	// TODO: there's some redundancy here with the object reference
	Object     *resolv.Object `json:"-"`
	IsOnGround bool           `json:"isOnGround"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Velocity struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Equal returns true if the player state is equal to the other player state
func (p *PlayerState) Equal(other *PlayerState) bool {
	return p.Position.X == other.Position.X &&
		p.Position.Y == other.Position.Y &&
		p.Velocity.X == other.Velocity.X &&
		p.Velocity.Y == other.Velocity.Y &&
		p.IsOnGround == other.IsOnGround
}

// Copy returns a copy of the player state with an empty object reference
func (p *PlayerState) Copy() *PlayerState {
	return &PlayerState{
		LastProcessedTimestamp: p.LastProcessedTimestamp,
		Position:               p.Position,
		Velocity:               p.Velocity,
		IsOnGround:             p.IsOnGround,
	}
}

type ConnectPlayerEvent struct {
	ClientID    uint32
	PlayerState *PlayerState
}

type DisconnectPlayerEvent struct {
	ClientID uint32
}
