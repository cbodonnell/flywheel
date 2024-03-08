package types

type GameState struct {
	// Timestamp is the time at which the game state was generated
	Timestamp int64 `json:"timestamp"`
	// Players maps client IDs to player states
	Players map[uint32]*PlayerState `json:"players"`
}

type PlayerState struct {
	Position Position `json:"position"`
	Velocity Velocity `json:"velocity"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Velocity struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type ConnectPlayerEvent struct {
	ClientID    uint32
	PlayerState *PlayerState
}

type DisconnectPlayerEvent struct {
	ClientID uint32
}
