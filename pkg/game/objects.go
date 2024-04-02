package game

import "github.com/cbodonnell/flywheel/pkg/game/types"

type ServerObject interface {
	Update(deltaTime float64)
}

type ServerObjects struct {
	npcs map[uint32]*NPCObject
}

func NewServerObjects() *ServerObjects {
	return &ServerObjects{
		npcs: make(map[uint32]*NPCObject),
	}
}

type NPCObject struct {
	ID    uint32
	State *types.NPCState
}

func (n *NPCObject) Update(deltaTime float64) {
	// TODO: implement
}
