package flow

type GameMode int

const (
	GameModeMenu GameMode = iota
	GameModePlay
	GameModeOver
	GameModeNetworkError
)

func (m GameMode) String() string {
	switch m {
	case GameModeMenu:
		return "Menu"
	case GameModePlay:
		return "Play"
	case GameModeOver:
		return "Over"
	case GameModeNetworkError:
		return "Network Error"
	}
	return "Unknown"
}
