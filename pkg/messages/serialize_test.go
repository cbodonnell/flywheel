package messages

import (
	"testing"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

func TestSerializeDeserializeGameState(t *testing.T) {
	type args struct {
		state *gametypes.GameState
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Basic game state",
			args: args{
				state: &gametypes.GameState{
					Timestamp: 1,
					Players: map[uint32]*gametypes.PlayerState{
						1: {
							LastProcessedTimestamp: 1,
							Position: gametypes.Position{
								X: 0.0,
								Y: 0.0,
							},
							Velocity: gametypes.Velocity{
								X: 0.0,
								Y: 0.0,
							},
							IsOnGround:    true,
							Animation:     gametypes.PlayerAnimationIdle,
							AnimationFlip: false,
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

			if !got.Players[1].Equal(tt.args.state.Players[1]) {
				t.Errorf("DeserializeGameState() = %v, want %v", got, tt.args.state)
			}
		})
	}
}
