package messages

import (
	"bytes"
	"fmt"
	"io"

	gamestatefb "github.com/cbodonnell/flywheel/flatbuffers/gamestate"
	messagefb "github.com/cbodonnell/flywheel/flatbuffers/message"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/klauspost/compress/zstd"
)

func SerializeMessage(m *Message) ([]byte, error) {
	b, err := SerializeMessageFlatbuffer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %v", err)
	}

	compressed := bytes.NewBuffer(nil)
	compWriter, err := zstd.NewWriter(compressed, zstd.WithEncoderLevel(zstd.SpeedFastest))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd writer: %v", err)
	}
	if _, err := compWriter.Write(b); err != nil {
		return nil, fmt.Errorf("failed to compress message: %v", err)
	}
	if err := compWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}

	return compressed.Bytes(), nil
}

func DeserializeMessage(data []byte) (*Message, error) {
	compReader, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd reader: %v", err)
	}
	b, err := io.ReadAll(compReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed message: %v", err)
	}

	message, err := DeserializeMessageFlatbuffer(b)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return message, nil
}

func SerializeMessageFlatbuffer(m *Message) ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)

	payload := builder.CreateByteVector(m.Payload)

	messagefb.MessageStart(builder)
	messagefb.MessageAddClientId(builder, m.ClientID)
	messagefb.MessageAddType(builder, byte(m.Type))
	messagefb.MessageAddPayload(builder, payload)
	messageOffset := messagefb.MessageEnd(builder)
	builder.Finish(messageOffset)
	b := builder.FinishedBytes()

	return b, nil
}

func DeserializeMessageFlatbuffer(b []byte) (*Message, error) {
	message := &Message{}
	messageFlatbuffer := messagefb.GetRootAsMessage(b, 0)
	message.ClientID = messageFlatbuffer.ClientId()
	message.Type = MessageType(messageFlatbuffer.Type())
	message.Payload = messageFlatbuffer.PayloadBytes()

	return message, nil
}

func SerializeGameState(state *ServerGameUpdate) ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)
	gameState := SerializeGameStateFlatbuffer(builder, state)
	builder.Finish(gameState)
	return builder.FinishedBytes(), nil
}

func DeserializeGameState(b []byte) (*ServerGameUpdate, error) {
	gameState, err := DeserializeGameStateFlatbuffer(b)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize game state: %v", err)
	}

	return gameState, nil
}

func SerializeGameStateFlatbuffer(builder *flatbuffers.Builder, state *ServerGameUpdate) flatbuffers.UOffsetT {
	playerStateKVs := make([]flatbuffers.UOffsetT, 0, len(state.Players))
	for k, v := range state.Players {
		playerState := SerializePlayerStateFlatbuffer(builder, v)

		gamestatefb.PlayerStateKeyValueStart(builder)
		gamestatefb.PlayerStateKeyValueAddKey(builder, k)
		gamestatefb.PlayerStateKeyValueAddValue(builder, playerState)
		playerStateKV := gamestatefb.PlayerStateKeyValueEnd(builder)

		playerStateKVs = append(playerStateKVs, playerStateKV)
	}
	gamestatefb.GameStateStartPlayersVector(builder, len(state.Players))
	for i := len(state.Players) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(playerStateKVs[i])
	}
	playerStates := builder.EndVector(len(state.Players))

	npcStateKVs := make([]flatbuffers.UOffsetT, 0, len(state.NPCs))
	for k, v := range state.NPCs {
		npcState := SerializeNPCStateFlatbuffer(builder, v)

		gamestatefb.NPCStateKeyValueStart(builder)
		gamestatefb.NPCStateKeyValueAddKey(builder, k)
		gamestatefb.NPCStateKeyValueAddValue(builder, npcState)
		npcStateKV := gamestatefb.NPCStateKeyValueEnd(builder)

		npcStateKVs = append(npcStateKVs, npcStateKV)
	}
	gamestatefb.GameStateStartNpcsVector(builder, len(state.NPCs))
	for i := len(state.NPCs) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(npcStateKVs[i])
	}
	npcStates := builder.EndVector(len(state.NPCs))

	gamestatefb.GameStateStart(builder)
	gamestatefb.GameStateAddTimestamp(builder, state.Timestamp)
	gamestatefb.GameStateAddPlayers(builder, playerStates)
	gamestatefb.GameStateAddNpcs(builder, npcStates)
	gameState := gamestatefb.GameStateEnd(builder)

	return gameState
}

func SerializePlayerStateFlatbuffer(builder *flatbuffers.Builder, state *PlayerStateUpdate) flatbuffers.UOffsetT {
	name := builder.CreateString(state.Name)

	gamestatefb.PositionStart(builder)
	gamestatefb.PositionAddX(builder, state.Position.X)
	gamestatefb.PositionAddY(builder, state.Position.Y)
	position := gamestatefb.PositionEnd(builder)

	gamestatefb.VelocityStart(builder)
	gamestatefb.VelocityAddX(builder, state.Velocity.X)
	gamestatefb.VelocityAddY(builder, state.Velocity.Y)
	velocity := gamestatefb.VelocityEnd(builder)

	gamestatefb.PlayerStateStart(builder)
	gamestatefb.PlayerStateAddLastProcessedTimestamp(builder, state.LastProcessedTimestamp)
	gamestatefb.PlayerStateAddCharacterId(builder, state.CharacterID)
	gamestatefb.PlayerStateAddName(builder, name)
	gamestatefb.PlayerStateAddPosition(builder, position)
	gamestatefb.PlayerStateAddVelocity(builder, velocity)
	gamestatefb.PlayerStateAddIsOnGround(builder, state.IsOnGround)
	gamestatefb.PlayerStateAddIsAttacking(builder, state.IsAttacking)
	gamestatefb.PlayerStateAddAnimation(builder, byte(state.Animation))
	gamestatefb.PlayerStateAddAnimationFlip(builder, state.AnimationFlip)
	gamestatefb.PlayerStateAddAnimationSequence(builder, state.AnimationSequence)
	gamestatefb.PlayerStateAddHitpoints(builder, state.Hitpoints)
	playerState := gamestatefb.PlayerStateEnd(builder)

	return playerState
}

func SerializeNPCStateFlatbuffer(builder *flatbuffers.Builder, state *NPCStateUpdate) flatbuffers.UOffsetT {
	gamestatefb.PositionStart(builder)
	gamestatefb.PositionAddX(builder, state.Position.X)
	gamestatefb.PositionAddY(builder, state.Position.Y)
	position := gamestatefb.PositionEnd(builder)

	gamestatefb.VelocityStart(builder)
	gamestatefb.VelocityAddX(builder, state.Velocity.X)
	gamestatefb.VelocityAddY(builder, state.Velocity.Y)
	velocity := gamestatefb.VelocityEnd(builder)

	gamestatefb.NPCStateStart(builder)
	gamestatefb.NPCStateAddPosition(builder, position)
	gamestatefb.NPCStateAddVelocity(builder, velocity)
	gamestatefb.NPCStateAddIsOnGround(builder, state.IsOnGround)
	gamestatefb.NPCStateAddAnimation(builder, byte(state.Animation))
	gamestatefb.NPCStateAddAnimationFlip(builder, state.AnimationFlip)
	gamestatefb.NPCStateAddAnimationSequence(builder, state.AnimationSequence)
	gamestatefb.NPCStateAddHitpoints(builder, state.Hitpoints)
	npcState := gamestatefb.NPCStateEnd(builder)

	return npcState
}

func DeserializeGameStateFlatbuffer(b []byte) (*ServerGameUpdate, error) {
	gameState := &ServerGameUpdate{}
	gameStateFlatbuffer := gamestatefb.GetRootAsGameState(b, 0)
	gameState.Timestamp = gameStateFlatbuffer.Timestamp()
	players := make(map[uint32]*PlayerStateUpdate)
	for i := 0; i < gameStateFlatbuffer.PlayersLength(); i++ {
		playerStateKV := &gamestatefb.PlayerStateKeyValue{}
		if !gameStateFlatbuffer.Players(playerStateKV, i) {
			return nil, fmt.Errorf("failed to get player state key value at index %d", i)
		}
		players[playerStateKV.Key()] = PlayerStateFlatbufferToPlayerStateUpdate(playerStateKV.Value(nil))
	}
	gameState.Players = players

	npcs := make(map[uint32]*NPCStateUpdate)
	for i := 0; i < gameStateFlatbuffer.NpcsLength(); i++ {
		npcStateKV := &gamestatefb.NPCStateKeyValue{}
		if !gameStateFlatbuffer.Npcs(npcStateKV, i) {
			return nil, fmt.Errorf("failed to get npc state key value at index %d", i)
		}
		npcs[npcStateKV.Key()] = NPCStateFlatbufferToNPCStateUpdate(npcStateKV.Value(nil))
	}
	gameState.NPCs = npcs

	return gameState, nil
}

func PlayerStateFlatbufferToPlayerStateUpdate(fb *gamestatefb.PlayerState) *PlayerStateUpdate {
	playerState := &PlayerStateUpdate{}
	playerState.LastProcessedTimestamp = fb.LastProcessedTimestamp()
	playerState.CharacterID = fb.CharacterId()
	playerState.Name = string(fb.Name())
	playerState.Position.X = fb.Position(nil).X()
	playerState.Position.Y = fb.Position(nil).Y()
	playerState.Velocity.X = fb.Velocity(nil).X()
	playerState.Velocity.Y = fb.Velocity(nil).Y()
	playerState.IsOnGround = fb.IsOnGround()
	playerState.IsAttacking = fb.IsAttacking()
	playerState.Animation = fb.Animation()
	playerState.AnimationFlip = fb.AnimationFlip()
	playerState.AnimationSequence = fb.AnimationSequence()
	playerState.Hitpoints = fb.Hitpoints()

	return playerState
}

func NPCStateFlatbufferToNPCStateUpdate(fb *gamestatefb.NPCState) *NPCStateUpdate {
	npcState := &NPCStateUpdate{}
	npcState.Position.X = fb.Position(nil).X()
	npcState.Position.Y = fb.Position(nil).Y()
	npcState.Velocity.X = fb.Velocity(nil).X()
	npcState.Velocity.Y = fb.Velocity(nil).Y()
	npcState.IsOnGround = fb.IsOnGround()
	npcState.Animation = fb.Animation()
	npcState.AnimationFlip = fb.AnimationFlip()
	npcState.AnimationSequence = fb.AnimationSequence()
	npcState.Hitpoints = fb.Hitpoints()

	return npcState
}

func SerializeServerPlayerUpdate(update *ServerPlayerUpdate) ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)
	playerState := SerializePlayerStateFlatbuffer(builder, update.PlayerState)
	gamestatefb.ServerPlayerUpdateStart(builder)
	gamestatefb.ServerPlayerUpdateAddTimestamp(builder, update.Timestamp)
	gamestatefb.ServerPlayerUpdateAddClientId(builder, update.ClientID)
	gamestatefb.ServerPlayerUpdateAddPlayerState(builder, playerState)
	serverPlayerUpdate := gamestatefb.ServerPlayerUpdateEnd(builder)
	builder.Finish(serverPlayerUpdate)
	return builder.FinishedBytes(), nil
}

func DeserializeServerPlayerUpdate(b []byte) (*ServerPlayerUpdate, error) {
	serverPlayerUpdate := &ServerPlayerUpdate{}
	serverPlayerUpdateFlatbuffer := gamestatefb.GetRootAsServerPlayerUpdate(b, 0)
	serverPlayerUpdate.Timestamp = serverPlayerUpdateFlatbuffer.Timestamp()
	serverPlayerUpdate.ClientID = serverPlayerUpdateFlatbuffer.ClientId()
	serverPlayerUpdate.PlayerState = PlayerStateFlatbufferToPlayerStateUpdate(serverPlayerUpdateFlatbuffer.PlayerState(nil))
	return serverPlayerUpdate, nil
}

func SerializeServerNPCUpdate(update *ServerNPCUpdate) ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)
	npcState := SerializeNPCStateFlatbuffer(builder, update.NPCState)
	gamestatefb.ServerNPCUpdateStart(builder)
	gamestatefb.ServerNPCUpdateAddTimestamp(builder, update.Timestamp)
	gamestatefb.ServerNPCUpdateAddNpcId(builder, update.NPCID)
	gamestatefb.ServerNPCUpdateAddNpcState(builder, npcState)
	serverNPCUpdate := gamestatefb.ServerNPCUpdateEnd(builder)
	builder.Finish(serverNPCUpdate)
	return builder.FinishedBytes(), nil
}

func DeserializeServerNPCUpdate(b []byte) (*ServerNPCUpdate, error) {
	serverNPCUpdate := &ServerNPCUpdate{}
	serverNPCUpdateFlatbuffer := gamestatefb.GetRootAsServerNPCUpdate(b, 0)
	serverNPCUpdate.Timestamp = serverNPCUpdateFlatbuffer.Timestamp()
	serverNPCUpdate.NPCID = serverNPCUpdateFlatbuffer.NpcId()
	serverNPCUpdate.NPCState = NPCStateFlatbufferToNPCStateUpdate(serverNPCUpdateFlatbuffer.NpcState(nil))
	return serverNPCUpdate, nil
}
