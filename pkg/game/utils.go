package game

import (
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/messages"
)

func ServerGameUpdateFromState(state *types.GameState) *messages.ServerGameUpdate {
	players := make(map[uint32]*messages.PlayerStateUpdate)
	for clientID, playerState := range state.Players {
		players[clientID] = PlayerStateUpdateFromState(playerState)
	}

	npcs := make(map[uint32]*messages.NPCStateUpdate)
	for npcID, npcState := range state.NPCs {
		npcs[npcID] = NPCStateUpdateFromState(npcState)
	}

	return &messages.ServerGameUpdate{
		Timestamp: state.Timestamp,
		Players:   players,
		NPCs:      npcs,
	}
}

func GameStateFromServerUpdate(update *messages.ServerGameUpdate) *types.GameState {
	players := make(map[uint32]*types.PlayerState)
	for clientID, playerState := range update.Players {
		players[clientID] = PlayerStateFromServerUpdate(playerState)
	}

	npcs := make(map[uint32]*types.NPCState)
	for npcID, npcState := range update.NPCs {
		npcs[npcID] = NPCStateFromServerUpdate(npcState)
	}

	return &types.GameState{
		Timestamp: update.Timestamp,
		Players:   players,
		NPCs:      npcs,
	}
}

func PlayerStateUpdateFromState(state *types.PlayerState) *messages.PlayerStateUpdate {
	return &messages.PlayerStateUpdate{
		LastProcessedTimestamp: state.LastProcessedTimestamp,
		UserID:                 state.UserID,
		Name:                   state.Name,
		Position:               state.Position,
		Velocity:               state.Velocity,
		IsOnGround:             state.IsOnGround,
		IsAttacking:            state.IsAttacking,
		Animation:              uint8(state.Animation),
		AnimationFlip:          state.AnimationFlip,
		AnimationSequence:      state.AnimationSequence,
		Hitpoints:              state.Hitpoints,
	}
}

func PlayerStateFromServerUpdate(update *messages.PlayerStateUpdate) *types.PlayerState {
	return &types.PlayerState{
		LastProcessedTimestamp: update.LastProcessedTimestamp,
		UserID:                 update.UserID,
		Name:                   update.Name,
		Position:               update.Position,
		Velocity:               update.Velocity,
		IsOnGround:             update.IsOnGround,
		IsAttacking:            update.IsAttacking,
		Animation:              types.PlayerAnimation(update.Animation),
		AnimationFlip:          update.AnimationFlip,
		AnimationSequence:      update.AnimationSequence,
		Hitpoints:              update.Hitpoints,
	}
}

func NPCStateUpdateFromState(state *types.NPCState) *messages.NPCStateUpdate {
	return &messages.NPCStateUpdate{
		Position:          state.Position,
		Velocity:          state.Velocity,
		IsOnGround:        state.IsOnGround,
		Animation:         uint8(state.Animation),
		AnimationFlip:     state.AnimationFlip,
		AnimationSequence: state.AnimationSequence,
		Hitpoints:         state.Hitpoints,
	}
}

func NPCStateFromServerUpdate(update *messages.NPCStateUpdate) *types.NPCState {
	return &types.NPCState{
		Position:          update.Position,
		Velocity:          update.Velocity,
		IsOnGround:        update.IsOnGround,
		Animation:         types.NPCAnimation(update.Animation),
		AnimationFlip:     update.AnimationFlip,
		AnimationSequence: update.AnimationSequence,
		Hitpoints:         update.Hitpoints,
	}
}
