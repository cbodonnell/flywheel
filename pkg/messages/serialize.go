package messages

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	gamestatefb "github.com/cbodonnell/flywheel/pkg/messages/flatbuffers/gamestate"
	messagefb "github.com/cbodonnell/flywheel/pkg/messages/flatbuffers/message"
	flatbuffers "github.com/google/flatbuffers/go"
)

func SerializeMessage(m *Message) ([]byte, error) {
	b, err := SerializeMessageFlatbuffer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %v", err)
	}

	// b, err := json.Marshal(m)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal message: %v", err)
	// }

	compressed := bytes.NewBuffer(nil)
	zlibWriter := zlib.NewWriter(compressed)
	if _, err := zlibWriter.Write(b); err != nil {
		return nil, fmt.Errorf("failed to compress message: %v", err)
	}
	if err := zlibWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}

	return compressed.Bytes(), nil
}

func DeserializeMessage(data []byte) (*Message, error) {
	zlibReader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decompress message: %v", err)
	}
	defer zlibReader.Close()

	b, err := io.ReadAll(zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed message: %v", err)
	}

	message, err := DeserializeMessageFlatbuffer(b)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	// message := &Message{}
	// if err := json.Unmarshal(b, message); err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	// }

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

func SerializeGameState(state *gametypes.GameState) ([]byte, error) {
	b, err := SerializeGameStateFlatbuffer(state)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize game state: %v", err)
	}

	// b, err := json.Marshal(state)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal game state: %v", err)
	// }

	return b, nil
}

func DeserializeGameState(b []byte) (*gametypes.GameState, error) {
	gameState, err := DeserializeGameStateFlatbuffer(b)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize game state: %v", err)
	}

	// gameState := &gametypes.GameState{}
	// if err := json.Unmarshal(b, gameState); err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal game state: %v", err)
	// }

	return gameState, nil
}

func SerializeGameStateFlatbuffer(state *gametypes.GameState) ([]byte, error) {
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

	gamestatefb.GameStateStart(builder)
	gamestatefb.GameStateAddTimestamp(builder, state.Timestamp)
	gamestatefb.GameStateAddPlayers(builder, playerStates)
	gamestateOffset := gamestatefb.GameStateEnd(builder)
	builder.Finish(gamestateOffset)
	b := builder.FinishedBytes()

	return b, nil
}

func DeserializeGameStateFlatbuffer(b []byte) (*gametypes.GameState, error) {
	gameState := &gametypes.GameState{}
	gameStateFlatbuffer := gamestatefb.GetRootAsGameState(b, 0)
	gameState.Timestamp = gameStateFlatbuffer.Timestamp()
	players := make(map[uint32]*gametypes.PlayerState)
	for i := 0; i < gameStateFlatbuffer.PlayersLength(); i++ {
		playerStateKV := &gamestatefb.PlayerStateKeyValue{}
		if !gameStateFlatbuffer.Players(playerStateKV, i) {
			return nil, fmt.Errorf("failed to get player state key value at index %d", i)
		}

		playerState := &gametypes.PlayerState{}
		playerState.LastProcessedTimestamp = playerStateKV.Value(nil).LastProcessedTimestamp()
		playerState.Position.X = playerStateKV.Value(nil).Position(nil).X()
		playerState.Position.Y = playerStateKV.Value(nil).Position(nil).Y()
		playerState.Velocity.X = playerStateKV.Value(nil).Velocity(nil).X()
		playerState.Velocity.Y = playerStateKV.Value(nil).Velocity(nil).Y()
		playerState.IsOnGround = playerStateKV.Value(nil).IsOnGround()
		playerState.Animation = gametypes.PlayerAnimation(playerStateKV.Value(nil).Animation())
		playerState.AnimationFlip = playerStateKV.Value(nil).AnimationFlip()

		players[playerStateKV.Key()] = playerState
	}
	gameState.Players = players

	return gameState, nil
}
