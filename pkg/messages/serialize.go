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
	compWriter, err := zstd.NewWriter(compressed)
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
	b, err := SerializeGameStateFlatbuffer(state)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize game state: %v", err)
	}

	return b, nil
}

func DeserializeGameState(b []byte) (*ServerGameUpdate, error) {
	gameState, err := DeserializeGameStateFlatbuffer(b)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize game state: %v", err)
	}

	return gameState, nil
}

func SerializeGameStateFlatbuffer(state *ServerGameUpdate) ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)

	playerStateKVs := make([]flatbuffers.UOffsetT, 0, len(state.Players))
	for k, v := range state.Players {
		gamestatefb.PositionStart(builder)
		gamestatefb.PositionAddX(builder, v.Position.X)
		gamestatefb.PositionAddY(builder, v.Position.Y)
		position := gamestatefb.PositionEnd(builder)

		gamestatefb.VelocityStart(builder)
		gamestatefb.VelocityAddX(builder, v.Velocity.X)
		gamestatefb.VelocityAddY(builder, v.Velocity.Y)
		velocity := gamestatefb.VelocityEnd(builder)

		gamestatefb.PlayerStateStart(builder)
		gamestatefb.PlayerStateAddLastProcessedTimestamp(builder, v.LastProcessedTimestamp)
		gamestatefb.PlayerStateAddPosition(builder, position)
		gamestatefb.PlayerStateAddVelocity(builder, velocity)
		gamestatefb.PlayerStateAddIsOnGround(builder, v.IsOnGround)
		gamestatefb.PlayerStateAddAnimation(builder, byte(v.Animation))
		gamestatefb.PlayerStateAddAnimationFlip(builder, v.AnimationFlip)
		playerState := gamestatefb.PlayerStateEnd(builder)

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
		gamestatefb.PositionStart(builder)
		gamestatefb.PositionAddX(builder, v.Position.X)
		gamestatefb.PositionAddY(builder, v.Position.Y)
		position := gamestatefb.PositionEnd(builder)

		gamestatefb.VelocityStart(builder)
		gamestatefb.VelocityAddX(builder, v.Velocity.X)
		gamestatefb.VelocityAddY(builder, v.Velocity.Y)
		velocity := gamestatefb.VelocityEnd(builder)

		gamestatefb.NPCStateStart(builder)
		gamestatefb.NPCStateAddPosition(builder, position)
		gamestatefb.NPCStateAddVelocity(builder, velocity)
		gamestatefb.NPCStateAddIsOnGround(builder, v.IsOnGround)
		npcState := gamestatefb.NPCStateEnd(builder)

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
	gamestateOffset := gamestatefb.GameStateEnd(builder)
	builder.Finish(gamestateOffset)
	b := builder.FinishedBytes()

	return b, nil
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

		playerState := &PlayerStateUpdate{}
		playerState.LastProcessedTimestamp = playerStateKV.Value(nil).LastProcessedTimestamp()
		playerState.Position.X = playerStateKV.Value(nil).Position(nil).X()
		playerState.Position.Y = playerStateKV.Value(nil).Position(nil).Y()
		playerState.Velocity.X = playerStateKV.Value(nil).Velocity(nil).X()
		playerState.Velocity.Y = playerStateKV.Value(nil).Velocity(nil).Y()
		playerState.IsOnGround = playerStateKV.Value(nil).IsOnGround()
		playerState.Animation = playerStateKV.Value(nil).Animation()
		playerState.AnimationFlip = playerStateKV.Value(nil).AnimationFlip()

		players[playerStateKV.Key()] = playerState
	}
	gameState.Players = players

	npcs := make(map[uint32]*NPCStateUpdate)
	for i := 0; i < gameStateFlatbuffer.NpcsLength(); i++ {
		npcStateKV := &gamestatefb.NPCStateKeyValue{}
		if !gameStateFlatbuffer.Npcs(npcStateKV, i) {
			return nil, fmt.Errorf("failed to get npc state key value at index %d", i)
		}

		npcState := &NPCStateUpdate{}
		npcState.Position.X = npcStateKV.Value(nil).Position(nil).X()
		npcState.Position.Y = npcStateKV.Value(nil).Position(nil).Y()
		npcState.Velocity.X = npcStateKV.Value(nil).Velocity(nil).X()
		npcState.Velocity.Y = npcStateKV.Value(nil).Velocity(nil).Y()
		npcState.IsOnGround = npcStateKV.Value(nil).IsOnGround()

		npcs[npcStateKV.Key()] = npcState
	}
	gameState.NPCs = npcs

	return gameState, nil
}
