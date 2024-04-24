package game

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cbodonnell/flywheel/client/input"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/client/scenes"
	authhandlers "github.com/cbodonnell/flywheel/pkg/auth/handlers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game implements ebiten.Game interface, which has Update, Draw and Layout methods.
type Game struct {
	// debug is a boolean value indicating whether debug mode is enabled.
	debug bool
	// authURL is the URL of the authentication server.
	authURL string
	// networkManager is the network manager.
	networkManager *network.NetworkManager
	// mode is the current game mode.
	mode GameMode
	// scene is the current scene.
	scene scenes.Scene
}

const (
	DefaultAuthServerURL = "http://localhost:8080"
)

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

type NewGameOptions struct {
	Debug          bool
	AuthURL        string
	NetworkManager *network.NetworkManager
}

func NewGame(opts NewGameOptions) (ebiten.Game, error) {
	g := &Game{
		debug:          opts.Debug,
		authURL:        opts.AuthURL,
		networkManager: opts.NetworkManager,
	}

	if err := g.loadMenu(); err != nil {
		return nil, fmt.Errorf("failed to load menu scene: %v", err)
	}

	return g, nil
}

func (g *Game) SetScene(scene scenes.Scene) error {
	if g.scene != nil {
		if err := g.scene.Destroy(); err != nil {
			return fmt.Errorf("failed to destroy previous scene: %v", err)
		}
	}

	g.scene = scene
	if err := g.scene.Init(); err != nil {
		return fmt.Errorf("failed to initialize scene: %v", err)
	}

	return nil
}

func (g *Game) loadMenu() error {
	menu, err := scenes.NewMenuScene(scenes.MenuSceneOptions{
		OnLogin: func(email, password string) {
			if err := g.login(email, password); err != nil {
				log.Error("Failed to start game: %v", err)
			}
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create menu scene: %v", err)
	}
	if err := g.SetScene(menu); err != nil {
		return fmt.Errorf("failed to set menu scene: %v", err)
	}
	g.mode = GameModeMenu
	return nil
}

func (g *Game) login(email, password string) error {
	idToken, err := g.getIDToken(email, password)
	if err != nil {
		return fmt.Errorf("failed to get ID token: %v", err)
	}

	if err := g.networkManager.Start(idToken); err != nil {
		log.Error("Failed to start network manager: %v", err)
		if err := g.loadNetworkError(); err != nil {
			return fmt.Errorf("failed to load network error scene: %v", err)
		}
		return nil
	}

	if err := g.loadGame(); err != nil {
		return fmt.Errorf("failed to load game scene: %v", err)
	}

	return nil
}

func (g *Game) getIDToken(email, password string) (string, error) {
	values := url.Values{}
	values.Set("email", email)
	values.Set("password", password)
	requestBody := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", g.authURL+"/login", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to login: status: %s, body: %s", resp.Status, string(b))
	}

	loginResponse := &authhandlers.LoginResponseBody{}
	if err := json.NewDecoder(resp.Body).Decode(loginResponse); err != nil {
		return "", fmt.Errorf("failed to decode login response: %v", err)
	}

	return loginResponse.IDToken, nil
}

func (g *Game) loadGame() error {
	gameScene, err := scenes.NewGameScene(g.networkManager)
	if err != nil {
		return fmt.Errorf("failed to create game scene: %v", err)
	}
	if err := g.SetScene(gameScene); err != nil {
		return fmt.Errorf("failed to set game scene: %v", err)
	}
	g.mode = GameModePlay
	return nil

}

func (g *Game) loadGameOver() error {
	g.networkManager.Stop()
	gameOver, err := scenes.NewGameOverScene()
	if err != nil {
		return fmt.Errorf("failed to create game over scene: %v", err)
	}
	if err := g.SetScene(gameOver); err != nil {
		return fmt.Errorf("failed to set game over scene: %v", err)
	}
	g.mode = GameModeOver
	return nil
}

func (g *Game) loadNetworkError() error {
	g.networkManager.Stop()
	networkError, err := scenes.NewErrorScene("Network Error")
	if err != nil {
		return fmt.Errorf("failed to create network error scene: %v", err)
	}
	if err := g.SetScene(networkError); err != nil {
		return fmt.Errorf("failed to set network error scene: %v", err)
	}
	g.mode = GameModeNetworkError
	return nil
}

func (g *Game) Update() error {
	// Update the network manager
	if err := g.networkManagerUpdate(); err != nil {
		return fmt.Errorf("failed to update network manager: %v", err)
	}

	// Handle input
	if err := g.handleInput(); err != nil {
		return fmt.Errorf("failed to handle input: %v", err)
	}

	// Update the current scene
	if err := g.scene.Update(); err != nil {
		return fmt.Errorf("failed to update scene: %v", err)
	}

	return nil
}

func (g *Game) networkManagerUpdate() error {
	if !g.networkManager.IsConnected() {
		return nil
	}

	g.networkManager.UpdateServerTime(1.0 / float64(ebiten.TPS()))

	if err := g.checkNetworkManagerErrors(); err != nil {
		log.Error("Network manager error: %v", err)
		if err := g.loadNetworkError(); err != nil {
			return fmt.Errorf("failed to load network error scene: %v", err)
		}
	}

	return nil
}

// checkNetworkManagerErrors checks the network manager for errors and returns any that are found.
func (g *Game) checkNetworkManagerErrors() error {
	select {
	case err := <-g.networkManager.TCPClientErrChan():
		return fmt.Errorf("TCP client error: %v", err)
	case err := <-g.networkManager.UDPClientErrChan():
		return fmt.Errorf("UDP client error: %v", err)
	default:
		return nil
	}
}

func (g *Game) handleInput() error {
	switch g.mode {
	case GameModeMenu:
		// if input.IsPositiveJustPressed() {
		// 	if err := g.loadGame(); err != nil {
		// 		return fmt.Errorf("failed to load game scene: %v", err)
		// 	}
		// }
	case GameModePlay:
		if input.IsNegativeJustPressed() {
			if err := g.loadGameOver(); err != nil {
				return fmt.Errorf("failed to load game over scene: %v", err)
			}
			break
		}
	case GameModeOver:
		if input.IsPositiveJustPressed() {
			if err := g.loadMenu(); err != nil {
				return fmt.Errorf("failed to load menu scene: %v", err)
			}
		}
	case GameModeNetworkError:
		if input.IsPositiveJustPressed() {
			if err := g.loadMenu(); err != nil {
				return fmt.Errorf("failed to load menu scene: %v", err)
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
	if g.debug {
		g.drawDebugOverlay(screen)
	}
}

func (g *Game) drawDebugOverlay(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n   FPS: %0.1f", ebiten.ActualFPS()))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n   TPS: %0.1f", ebiten.ActualTPS()))

	if !g.networkManager.IsConnected() {
		return
	}

	serverSettings := g.networkManager.ServerSettings()
	serverTime, ping := g.networkManager.ServerTime()
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n   Server: %s", serverSettings.Hostname))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n   Ping: %0.1f", ping))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n\n   Time: %0.0f", serverTime))
}

const (
	DefaultScreenWidth  = 640
	DefaultScreenHeight = 480
)

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DefaultScreenWidth, DefaultScreenHeight
}
