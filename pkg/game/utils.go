package game

import (
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/messages"
)

func ServerGameUpdateFromState(state *types.GameState) *messages.ServerGameUpdate {
	players := make(map[uint32]*messages.PlayerStateUpdate)
	for clientID, playerState := range state.Players {
		players[clientID] = &messages.PlayerStateUpdate{
			LastProcessedTimestamp: playerState.LastProcessedTimestamp,
			Position:               playerState.Position,
			Velocity:               playerState.Velocity,
			IsOnGround:             playerState.IsOnGround,
			Animation:              uint8(playerState.Animation),
			AnimationFlip:          playerState.AnimationFlip,
		}
	}

	npcs := make(map[uint32]*messages.NPCStateUpdate)
	for npcID, npcState := range state.NPCs {
		if !npcState.Exists() {
			// don't send updates for NPCs that aren't in the game
			continue
		}
		npcs[npcID] = &messages.NPCStateUpdate{
			Position:   npcState.Position,
			Velocity:   npcState.Velocity,
			IsOnGround: npcState.IsOnGround,
		}
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
		players[clientID] = &types.PlayerState{
			LastProcessedTimestamp: playerState.LastProcessedTimestamp,
			Position:               playerState.Position,
			Velocity:               playerState.Velocity,
			IsOnGround:             playerState.IsOnGround,
			Animation:              types.PlayerAnimation(playerState.Animation),
			AnimationFlip:          playerState.AnimationFlip,
		}
	}

	npcs := make(map[uint32]*types.NPCState)
	for npcID, npcState := range update.NPCs {
		npcs[npcID] = &types.NPCState{
			Position:   npcState.Position,
			Velocity:   npcState.Velocity,
			IsOnGround: npcState.IsOnGround,
		}
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
		Position:               state.Position,
		Velocity:               state.Velocity,
		IsOnGround:             state.IsOnGround,
		Animation:              uint8(state.Animation),
		AnimationFlip:          state.AnimationFlip,
	}
}

func PlayerStateFromServerUpdate(update *messages.PlayerStateUpdate) *types.PlayerState {
	return &types.PlayerState{
		LastProcessedTimestamp: update.LastProcessedTimestamp,
		Position:               update.Position,
		Velocity:               update.Velocity,
		IsOnGround:             update.IsOnGround,
		Animation:              types.PlayerAnimation(update.Animation),
		AnimationFlip:          update.AnimationFlip,
	}
}

func NPCStateUpdateFromState(state *types.NPCState) *messages.NPCStateUpdate {
	return &messages.NPCStateUpdate{
		Position:   state.Position,
		Velocity:   state.Velocity,
		IsOnGround: state.IsOnGround,
	}
}

func NPCStateFromServerUpdate(update *messages.NPCStateUpdate) *types.NPCState {
	return &types.NPCState{
		Position:   update.Position,
		Velocity:   update.Velocity,
		IsOnGround: update.IsOnGround,
	}
}
