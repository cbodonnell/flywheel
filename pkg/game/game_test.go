package game

import (
	"encoding/json"
	"fmt"
	"testing"

	mocks "github.com/cbodonnell/flywheel/mocks/github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/solarlune/resolv"
	"github.com/stretchr/testify/assert"
)

func TestGameManager_processClientMessages(t *testing.T) {
	mockQueue := mocks.NewQueue(t)

	type fields struct {
		clientMessageQueue queue.Queue
		gameState          *types.GameState
	}
	tests := []struct {
		name   string
		fields fields
		setup  func()
		want   *types.GameState
	}{
		{
			name: "basic movement",
			fields: fields{
				clientMessageQueue: mockQueue,
				gameState: &types.GameState{
					Players: map[uint32]*types.PlayerState{
						1: {
							PlayerID: "player-1",
							Position: kinematic.Vector{
								X: 0,
								Y: 0,
							},
							Velocity: kinematic.Vector{
								X: 0,
								Y: 0,
							},
							Object: resolv.NewObject(constants.PlayerStartingX, constants.PlayerStartingY, constants.PlayerWidth, constants.PlayerHeight, types.CollisionSpaceTagPlayer),
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
						Timestamp: 2,
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

				mockQueue.EXPECT().ReadAllMessages().Return(testMessages, nil).Once()
			},
			want: &types.GameState{
				Players: map[uint32]*types.PlayerState{
					1: {
						PlayerID: "player-1",
						Position: kinematic.Vector{
							X: 35,
							Y: -58.80000000000001,
						},
						Velocity: kinematic.Vector{
							X: 350,
							Y: -588,
						},
					},
				},
			},
		},
		{
			name: "no messages",
			fields: fields{
				clientMessageQueue: mockQueue,
				gameState: &types.GameState{
					Players: map[uint32]*types.PlayerState{
						1: {
							PlayerID: "player-1",
							Position: kinematic.Vector{
								X: 0,
								Y: 0,
							},
							Velocity: kinematic.Vector{
								X: 0,
								Y: 0,
							},
							Object: resolv.NewObject(constants.PlayerStartingX, constants.PlayerStartingY, constants.PlayerWidth, constants.PlayerHeight, types.CollisionSpaceTagPlayer),
						},
					},
				},
			},
			setup: func() {
				mockQueue.EXPECT().ReadAllMessages().Return([]interface{}{}, nil).Once()
			},
			want: &types.GameState{
				Players: map[uint32]*types.PlayerState{
					1: {
						PlayerID: "player-1",
						Position: kinematic.Vector{
							X: 0,
							Y: 0,
						},
						Velocity: kinematic.Vector{
							X: 0,
							Y: 0,
						},
					},
				},
			},
		},
		{
			name: "out of order messages",
			fields: fields{
				clientMessageQueue: mockQueue,
				gameState: &types.GameState{
					Players: map[uint32]*types.PlayerState{
						1: {
							PlayerID: "player-1",
							Position: kinematic.Vector{
								X: 0,
								Y: 0,
							},
							Velocity: kinematic.Vector{
								X: 0,
								Y: 0,
							},
							Object: resolv.NewObject(constants.PlayerStartingX, constants.PlayerStartingY, constants.PlayerWidth, constants.PlayerHeight, types.CollisionSpaceTagPlayer),
						},
					},
				},
			},
			setup: func() {
				testClientPlayerUpdates := []messages.ClientPlayerUpdate{
					{
						Timestamp: 2,
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

				mockQueue.EXPECT().ReadAllMessages().Return(testMessages, nil).Once()
			},
			want: &types.GameState{
				Players: map[uint32]*types.PlayerState{
					1: {
						PlayerID: "player-1",
						Position: kinematic.Vector{
							X: 0.0,
							Y: -14.700000000000003,
						},
						Velocity: kinematic.Vector{
							X: 0,
							Y: -294,
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
				clientMessageQueue: tt.fields.clientMessageQueue,
				gameState:          tt.fields.gameState,
			}
			gm.processClientMessages()
			if tt.want != nil {
				for clientID, wantPlayerState := range tt.want.Players {
					assert.Equal(t, wantPlayerState.Position, tt.fields.gameState.Players[clientID].Position, fmt.Sprintf("Position for clientID %d", clientID))
					assert.Equal(t, wantPlayerState.Velocity, tt.fields.gameState.Players[clientID].Velocity, fmt.Sprintf("Velocity for clientID %d", clientID))
				}
			}
		})
	}
}
