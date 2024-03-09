package game

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	mocks "github.com/cbodonnell/flywheel/mocks/github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/state"
	"github.com/cbodonnell/flywheel/pkg/workers"
	"github.com/solarlune/resolv"
	"github.com/stretchr/testify/assert"
)

func TestGameManager_processClientMessages(t *testing.T) {
	mockQueue := mocks.NewQueue(t)

	type fields struct {
		clientManager        *clients.ClientManager
		clientMessageQueue   queue.Queue
		connectionEventQueue queue.Queue
		repository           repositories.Repository
		stateManager         state.StateManager
		savePlayerStateChan  chan<- workers.SavePlayerStateRequest
		gameLoopInterval     time.Duration
		collisionSpace       *resolv.Space
	}
	type args struct {
		gameState *types.GameState
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   *types.GameState
	}{
		{
			name: "basic movement",
			fields: fields{
				clientMessageQueue: mockQueue,
			},
			args: args{
				gameState: &types.GameState{
					Players: map[uint32]*types.PlayerState{
						1: {
							Position: types.Position{
								X: 0,
								Y: 0,
							},
							Velocity: types.Velocity{
								X: 0,
								Y: 0,
							},
							Object: resolv.NewObject(constants.PlayerStartingX, constants.PlayerStartingY, constants.PlayerWidth, constants.PlayerHeight),
						},
					},
				},
			},
			setup: func() {
				testClientPlayerUpdates := []messages.ClientPlayerUpdate{
					{
						Timestamp: 1,
						InputX:    0,
						InputY:    0,
						DeltaTime: 0.1,
					},
					{
						Timestamp: 1,
						InputX:    1,
						InputY:    0,
						DeltaTime: 0.1,
					},
				}
				testMessages := make([]interface{}, len(testClientPlayerUpdates))

				for i, update := range testClientPlayerUpdates {
					payload, err := json.Marshal(update)
					if err != nil {
						t.Fatalf("failed to marshal payload: %v", err)
					}
					testMessages[i] = &messages.Message{
						ClientID: 1,
						Type:     messages.MessageTypeClientPlayerUpdate,
						Payload:  payload,
					}
				}

				mockQueue.EXPECT().ReadAllMessages().Return(testMessages, nil)
			},
			want: &types.GameState{
				Players: map[uint32]*types.PlayerState{
					1: {
						Position: types.Position{
							X: 0.5,
							Y: -0.19600000000000004,
						},
						Velocity: types.Velocity{
							X: 5,
							Y: -1.9600000000000002,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			gm := &GameManager{
				clientManager:        tt.fields.clientManager,
				clientMessageQueue:   tt.fields.clientMessageQueue,
				connectionEventQueue: tt.fields.connectionEventQueue,
				repository:           tt.fields.repository,
				stateManager:         tt.fields.stateManager,
				savePlayerStateChan:  tt.fields.savePlayerStateChan,
				gameLoopInterval:     tt.fields.gameLoopInterval,
				collisionSpace:       tt.fields.collisionSpace,
			}
			gm.processClientMessages(tt.args.gameState)
			if tt.want != nil {
				for clientID, wantPlayerState := range tt.want.Players {
					assert.Equal(t, wantPlayerState.Position, tt.args.gameState.Players[clientID].Position, fmt.Sprintf("Position for clientID %d", clientID))
					assert.Equal(t, wantPlayerState.Velocity, tt.args.gameState.Players[clientID].Velocity, fmt.Sprintf("Velocity for clientID %d", clientID))
				}
			}
		})
	}
}
