package game

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cbodonnell/flywheel/client/input"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/client/scenes"
	"github.com/cbodonnell/flywheel/client/ui"
	authhandlers "github.com/cbodonnell/flywheel/pkg/auth/handlers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories/models"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game implements ebiten.Game interface, which has Update, Draw and Layout methods.
type Game struct {
	// debug is a boolean value indicating whether debug mode is enabled.
	debug bool
	// TODO: make auth and api clients to pass in
	// auth is the game authentication configuration.
	auth GameAuth
	// api is the game API configuration.
	api GameAPI
	// networkManager is the network manager.
	networkManager *network.NetworkManager
	// gameAutomation is the game automation configuration.
	gameAutomation *GameAutomation
	// mode is the current game mode.
	mode GameMode
	// scene is the current scene.
	scene scenes.Scene
}

type GameAuth struct {
	URL          string
	Email        string
	Password     string
	IDToken      string
	IDTokenExp   time.Time
	RefreshToken string
}

type GameAPI struct {
	URL string
}

var (
	DefaultAuthServerURL = "http://localhost:8080"
	DefaultAPIServerURL  = "http://localhost:9090"
)

type GameMode int

const (
	GameModeAuth GameMode = iota
	GameModePlay
	GameModeOver
	GameModeNetworkError
)

func (m GameMode) String() string {
	switch m {
	case GameModeAuth:
		return "Auth"
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
	APIURL         string
	NetworkManager *network.NetworkManager
	GameAutomation *GameAutomation
}

type GameAutomation struct {
	Email    string
	Password string
}

func NewGame(opts NewGameOptions) (ebiten.Game, error) {
	g := &Game{
		debug: opts.Debug,
		auth: GameAuth{
			URL: opts.AuthURL,
		},
		api: GameAPI{
			URL: opts.APIURL,
		},
		networkManager: opts.NetworkManager,
		gameAutomation: opts.GameAutomation,
	}

	isAutomated := false
	if g.gameAutomation != nil {
		isAutomated = true
		g.setEmailPassword(g.gameAutomation.Email, g.gameAutomation.Password)
	}

	if err := g.loadAuth(isAutomated); err != nil {
		return nil, fmt.Errorf("failed to load auth scene: %v", err)
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

func (g *Game) loadAuth(isAutomated bool) error {
	if err := g.login(); err == nil {
		return nil
	} else if isAutomated {
		return fmt.Errorf("failed to login with automation: %v", err)
	}

	authSceneOpts := scenes.AuthSceneOptions{
		OnLogin: func(email string, password string) error {
			g.setEmailPassword(email, password)
			if err := g.login(); err != nil {
				if actionableErr, ok := err.(*ui.ActionableError); ok {
					return actionableErr
				}
				return fmt.Errorf("failed to login: %v", err)
			}
			return nil
		},
		OnRegister: func(email string, password string) error {
			g.setEmailPassword(email, password)
			if err := g.register(); err != nil {
				if actionableErr, ok := err.(*ui.ActionableError); ok {
					return actionableErr
				}
				return fmt.Errorf("failed to register: %v", err)
			}
			return nil
		},
	}
	auth, err := scenes.NewAuthScene(authSceneOpts)
	if err != nil {
		return fmt.Errorf("failed to create auth scene: %v", err)
	}
	if err := g.SetScene(auth); err != nil {
		return fmt.Errorf("failed to set auth scene: %v", err)
	}
	g.mode = GameModeAuth
	return nil
}

func (g *Game) login() error {
	if err := g.refreshIDToken(); err != nil {
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return actionableErr
		}
		return fmt.Errorf("failed to refresh ID token: %v", err)
	}

	if err := g.loadCharacterSelection(); err != nil {
		return fmt.Errorf("failed to load character selection scene: %v", err)
	}

	return nil
}

func (g *Game) register() error {
	if err := g.registerWithEmailPassword(); err != nil {
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return actionableErr
		}
		return fmt.Errorf("failed to refresh ID token: %v", err)
	}

	if err := g.loadCharacterSelection(); err != nil {
		return fmt.Errorf("failed to load character selection scene: %v", err)
	}

	return nil
}

func (g *Game) registerWithEmailPassword() error {
	values := url.Values{}
	values.Set("email", g.auth.Email)
	values.Set("password", g.auth.Password)
	requestBody := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/register", g.auth.URL), requestBody)
	if err != nil {
		return fmt.Errorf("failed to create register request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send register request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		msg := string(b)
		if len(msg) == 0 {
			msg = resp.Status
		}
		return &ui.ActionableError{Message: msg}
	}

	registerResponse := &authhandlers.RegisterResponseBody{}
	if err := json.NewDecoder(resp.Body).Decode(registerResponse); err != nil {
		return fmt.Errorf("failed to decode register response: %v", err)
	}

	g.auth.IDToken = registerResponse.IDToken
	expiresIn, err := strconv.Atoi(registerResponse.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to parse expires in: %v", err)
	}
	g.auth.IDTokenExp = time.Now().Add(time.Duration(expiresIn) * time.Second)
	g.auth.RefreshToken = registerResponse.RefreshToken

	return nil
}

func (g *Game) setEmailPassword(email, password string) {
	g.auth.Email = email
	g.auth.Password = password
}

func (g *Game) refreshIDToken() error {
	if g.auth.IDToken != "" {
		if time.Now().Before(g.auth.IDTokenExp) {
			// we have a valid token
			return nil
		} else if g.auth.RefreshToken != "" {
			log.Debug("Token expired, refreshing")
			// we have a refresh token
			if err := g.getIDTokenWithRefreshToken(); err == nil {
				return nil
			} else {
				// failed to refresh token, fall back to email/password
				log.Error("Failed to refresh token: %v", err)
			}
		}
	}

	if g.auth.Email == "" || g.auth.Password == "" {
		return fmt.Errorf("email and password are required")
	}

	log.Debug("Getting ID token with email/password")
	err := g.getIDTokenWithEmailPassword()
	if err != nil {
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return actionableErr
		}
		return fmt.Errorf("failed to get ID token: %v", err)
	}

	return nil
}

func (g *Game) getIDTokenWithEmailPassword() error {
	values := url.Values{}
	values.Set("email", g.auth.Email)
	values.Set("password", g.auth.Password)
	requestBody := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", g.auth.URL), requestBody)
	if err != nil {
		return fmt.Errorf("failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		msg := string(b)
		if len(msg) == 0 {
			msg = resp.Status
		}
		return &ui.ActionableError{Message: msg}
	}

	loginResponse := &authhandlers.LoginResponseBody{}
	if err := json.NewDecoder(resp.Body).Decode(loginResponse); err != nil {
		return fmt.Errorf("failed to decode login response: %v", err)
	}

	g.auth.IDToken = loginResponse.IDToken
	expiresIn, err := strconv.Atoi(loginResponse.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to parse expires in: %v", err)
	}
	g.auth.IDTokenExp = time.Now().Add(time.Duration(expiresIn) * time.Second)
	g.auth.RefreshToken = loginResponse.RefreshToken

	return nil
}

func (g *Game) getIDTokenWithRefreshToken() error {
	values := url.Values{}
	values.Set("refreshToken", g.auth.RefreshToken)
	requestBody := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/refresh", g.auth.URL), requestBody)
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send refresh request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to refresh token: status: %s, body: %s", resp.Status, string(b))
	}

	refreshResponse := &authhandlers.RefreshResponseBody{}
	if err := json.NewDecoder(resp.Body).Decode(refreshResponse); err != nil {
		return fmt.Errorf("failed to decode refresh response: %v", err)
	}

	g.auth.IDToken = refreshResponse.IDToken
	expiresIn, err := strconv.Atoi(refreshResponse.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to parse expires in: %v", err)
	}
	g.auth.IDTokenExp = time.Now().Add(time.Duration(expiresIn) * time.Second)
	g.auth.RefreshToken = refreshResponse.RefreshToken

	return nil
}

func (g *Game) loadCharacterSelection() error {
	characterSelectionOpts := scenes.CharacterSelectionSceneOpts{
		FetchCharacters:   g.fetchCharacters,
		CreateCharacter:   g.createCharacter,
		DeleteCharacter:   g.deleteCharacter,
		OnSelectCharacter: g.onSelectCharacter,
	}
	characterSelection, err := scenes.NewCharacterSelectionScene(characterSelectionOpts)
	if err != nil {
		return fmt.Errorf("failed to create character selection scene: %v", err)
	}
	if err := g.SetScene(characterSelection); err != nil {
		return fmt.Errorf("failed to set character selection scene: %v", err)
	}
	g.mode = GameModeAuth
	return nil
}

func (g *Game) onSelectCharacter(characterID int32) error {
	if err := g.refreshIDToken(); err != nil {
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return actionableErr
		}
		return fmt.Errorf("failed to refresh ID token: %v", err)
	}

	if err := g.networkManager.Start(g.auth.IDToken, characterID); err != nil {
		if err := g.networkManager.Stop(); err != nil {
			return fmt.Errorf("failed to stop network manager: %v", err)
		}
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return actionableErr
		}
		return fmt.Errorf("failed to start network manager: %v", err)
	}

	if err := g.loadGame(); err != nil {
		return fmt.Errorf("failed to load game scene: %v", err)
	}

	return nil
}

func (g *Game) fetchCharacters() ([]*models.Character, error) {
	if err := g.refreshIDToken(); err != nil {
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return nil, actionableErr
		}
		return nil, fmt.Errorf("failed to refresh ID token: %v", err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/characters", g.api.URL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create characters request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.auth.IDToken))

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send characters request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		msg := string(b)
		if len(msg) == 0 {
			msg = resp.Status
		}
		return nil, &ui.ActionableError{Message: msg}
	}

	characters := make([]*models.Character, 0)
	if err := json.NewDecoder(resp.Body).Decode(&characters); err != nil {
		return nil, fmt.Errorf("failed to decode characters response: %v", err)
	}

	return characters, nil
}

func (g *Game) createCharacter(name string) (*models.Character, error) {
	if err := g.refreshIDToken(); err != nil {
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return nil, actionableErr
		}
		return nil, fmt.Errorf("failed to refresh ID token: %v", err)
	}

	values := url.Values{}
	values.Set("name", name)
	requestBody := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/characters", g.api.URL), requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create create character request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.auth.IDToken))

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send create character request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		msg := string(b)
		if len(msg) == 0 {
			msg = resp.Status
		}
		return nil, &ui.ActionableError{Message: msg}
	}

	character := &models.Character{}
	if err := json.NewDecoder(resp.Body).Decode(character); err != nil {
		return nil, fmt.Errorf("failed to decode character response: %v", err)
	}

	return character, nil
}

func (g *Game) deleteCharacter(characterID int32) error {
	if err := g.refreshIDToken(); err != nil {
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			return actionableErr
		}
		return fmt.Errorf("failed to refresh ID token: %v", err)
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/characters/%d", g.api.URL, characterID), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete character request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.auth.IDToken))

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete character request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		msg := string(b)
		if len(msg) == 0 {
			msg = resp.Status
		}
		return &ui.ActionableError{Message: msg}
	}

	return nil

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
	case GameModeAuth:
		// no input handling here
	case GameModePlay:
		if input.IsNegativeJustPressed() {
			if err := g.loadGameOver(); err != nil {
				return fmt.Errorf("failed to load game over scene: %v", err)
			}
			break
		}
	case GameModeOver:
		if input.IsPositiveJustPressed() {
			if err := g.loadAuth(false); err != nil {
				return fmt.Errorf("failed to load auth scene: %v", err)
			}
		}
	case GameModeNetworkError:
		if input.IsPositiveJustPressed() {
			if err := g.loadAuth(false); err != nil {
				return fmt.Errorf("failed to load auth scene: %v", err)
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

	serverSettings := g.networkManager.ServerSettings()
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n   Game Server: %s", serverSettings.Hostname))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n   Auth Server: %s", g.auth.URL))

	if !g.networkManager.IsConnected() {
		return
	}

	serverTime, ping := g.networkManager.ServerTime()
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n\n   Ping: %0.1f", ping))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n\n\n   Time: %0.0f", serverTime))
}

const (
	DefaultScreenWidth  = 640
	DefaultScreenHeight = 480
)

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DefaultScreenWidth, DefaultScreenHeight
}
