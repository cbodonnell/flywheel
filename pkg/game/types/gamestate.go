package types

import "github.com/solarlune/resolv"

const (
	CollisionSpaceTagPlayer string = "player"
	CollisionSpaceTagNPC    string = "npc"
	CollisionSpaceTagLevel  string = "level"
)

type GameState struct {
	// Timestamp is the time at which the game state was generated
	Timestamp int64
	// Players maps client IDs to player states
	Players map[uint32]*PlayerState
	// NPCs maps enemy IDs to NPC states
	NPCs map[uint32]*NPCState
	// CollisionSpace is a resolv.Space used for collision detection
	CollisionSpace *resolv.Space
}

func NewGameState(collisionSpace *resolv.Space) *GameState {
	return &GameState{
		Timestamp:      0,
		Players:        make(map[uint32]*PlayerState),
		NPCs:           make(map[uint32]*NPCState),
		CollisionSpace: collisionSpace,
	}
}

func (g *GameState) Copy() *GameState {
	newGameState := &GameState{
		Timestamp: g.Timestamp,
		Players:   make(map[uint32]*PlayerState),
		NPCs:      make(map[uint32]*NPCState),
	}
	for id, player := range g.Players {
		newGameState.Players[id] = player.Copy()
	}
	for id, npc := range g.NPCs {
		newGameState.NPCs[id] = npc.Copy()
	}
	return newGameState
}

func (g *GameState) SetTimestamp(timestamp int64) {
	g.Timestamp = timestamp
}

func (g *GameState) AddPlayer(id uint32, state *PlayerState) {
	g.Players[id] = state
}

func (g *GameState) RemovePlayer(id uint32) {
	delete(g.Players, id)
}

func (g *GameState) AddNPC(id uint32, state *NPCState) {
	g.NPCs[id] = state
}

func (g *GameState) RemoveNPC(id uint32) {
	delete(g.NPCs, id)
}
