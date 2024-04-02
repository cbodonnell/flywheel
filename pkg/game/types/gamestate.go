package types

type GameState struct {
	// Timestamp is the time at which the game state was generated
	Timestamp int64 `json:"timestamp"`
	// Players maps client IDs to player states
	Players map[uint32]*PlayerState `json:"players"`
	// NPCs maps enemy IDs to NPC states
	NPCs map[uint32]*NPCState `json:"npcs"`
}
