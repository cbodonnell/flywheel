package messages

import (
	"testing"

	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/stretchr/testify/assert"
)

func TestSerializeDeserializeGameState(t *testing.T) {
	type args struct {
		state *ServerGameUpdate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Basic game state",
			args: args{
				state: &ServerGameUpdate{
					Timestamp: 1,
					Players: map[uint32]*PlayerStateUpdate{
						1: {
							LastProcessedTimestamp: 1,
							Position: kinematic.Vector{
								X: 0.0,
								Y: 0.0,
							},
							Velocity: kinematic.Vector{
								X: 0.0,
								Y: 0.0,
							},
							IsOnGround:    true,
							Animation:     0,
							AnimationFlip: false,
						},
					},
					NPCs: map[uint32]*NPCStateUpdate{
						1: {
							Position: kinematic.Vector{
								X: 0.0,
								Y: 0.0,
							},
							Velocity: kinematic.Vector{
								X: 0.0,
								Y: 0.0,
							},
							IsOnGround:    true,
							Animation:     0,
							AnimationFlip: false,
							Hitpoints:     100,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := SerializeGameState(tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeGameState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := DeserializeGameState(b)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeGameState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.args.state, got)
		})
	}
}
