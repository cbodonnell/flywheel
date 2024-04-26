package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/network"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/workers"
)

type GameManager struct {
	clientManager        *network.ClientManager
	clientMessageQueue   queue.Queue
	connectionEventQueue queue.Queue
	repository           repositories.Repository
	gameState            *types.GameState
	savePlayerStateChan  chan<- workers.SavePlayerStateRequest
	gameLoopInterval     time.Duration
}

// NewGameManagerOptions contains options for creating a new GameManager.
type NewGameManagerOptions struct {
	ClientManager        *network.ClientManager
	ClientMessageQueue   queue.Queue
	ConnectionEventQueue queue.Queue
	Repository           repositories.Repository
	GameState            *types.GameState
	SavePlayerStateChan  chan<- workers.SavePlayerStateRequest
	GameLoopInterval     time.Duration
}

func NewGameManager(opts NewGameManagerOptions) *GameManager {
	return &GameManager{
		clientManager:        opts.ClientManager,
		clientMessageQueue:   opts.ClientMessageQueue,
		connectionEventQueue: opts.ConnectionEventQueue,
		repository:           opts.Repository,
		gameState:            opts.GameState,
		savePlayerStateChan:  opts.SavePlayerStateChan,
		gameLoopInterval:     opts.GameLoopInterval,
	}
}

// Start starts the game loop.
func (gm *GameManager) Start(ctx context.Context) error {
	if err := gm.initializeGameState(ctx); err != nil {
		return fmt.Errorf("failed to initialize game state: %v", err)
	}

	ticker := time.NewTicker(gm.gameLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-ticker.C:
			err := gm.gameTick(ctx, t)
			if err != nil {
				log.Error("Failed to run game tick: %v", err)
			}
		}
	}
}

func (gm *GameManager) Stop() {
	// TODO: gracefully stop the game and save the game state
}

func (gm *GameManager) initializeGameState(_ context.Context) error {
	npcState := types.NewNPCState(constants.NPCStartingX, constants.NPCStartingY)
	gm.gameState.NPCs[1] = npcState
	gm.gameState.CollisionSpace.Add(npcState.Object)

	return nil
}

// gameTick runs one iteration of the game loop.
func (gm *GameManager) gameTick(_ context.Context, t time.Time) error {
	gm.gameState.Timestamp = t.UnixMilli()
	gm.processConnectionEvents()
	gm.processClientMessages()
	gm.updateServerObjects(gm.gameLoopInterval.Seconds())
	gm.broadcastGameState()

	return nil
}

// processConnectionEvents processes all pending connection events in the queue,
// updates the game state, and notifies connected clients
func (gm *GameManager) processConnectionEvents() {
	pendingEvents, err := gm.connectionEventQueue.ReadAllMessages()
	if err != nil {
		log.Error("Failed to read connection events: %v", err)
		return
	}
	for _, item := range pendingEvents {
		switch event := item.(type) {
		case *types.ConnectPlayerEvent:
			playerState := types.NewPlayerState(event.UserID, event.Position.X, event.Position.Y)
			log.Debug("Player %s created as %s", playerState.UserID, playerState.Name)
			// add the player to the game state
			gm.gameState.Players[event.ClientID] = playerState
			// add the player object to the collision space
			gm.gameState.CollisionSpace.Add(playerState.Object)

			// send a message to connected clients to add the player to the game
			playerConnect := &messages.ServerPlayerConnect{
				ClientID:    event.ClientID,
				PlayerState: PlayerStateUpdateFromState(playerState),
			}
			payload, err := json.Marshal(playerConnect)
			if err != nil {
				log.Error("Failed to marshal player state: %v", err)
				continue
			}

			for _, client := range gm.clientManager.GetClients() {
				msg := &messages.Message{
					ClientID: 0, // ClientID 0 means the message is from the server
					Type:     messages.MessageTypeServerPlayerConnect,
					Payload:  payload,
				}

				err := network.WriteMessageToTCP(client.TCPConn, msg)
				if err != nil {
					log.Error("Failed to write message to TCP connection for client %d: %v", client.ID, err)
					continue
				}
			}
		case *types.DisconnectPlayerEvent:
			// send a request to save the player state before deleting it
			saveRequest := workers.SavePlayerStateRequest{
				Timestamp:   gm.gameState.Timestamp,
				UserID:      gm.gameState.Players[event.ClientID].UserID,
				PlayerState: gm.gameState.Players[event.ClientID],
			}
			gm.savePlayerStateChan <- saveRequest
			// remove the player object from the collision space
			gm.gameState.CollisionSpace.Remove(gm.gameState.Players[event.ClientID].Object)
			// delete the player from the game state
			delete(gm.gameState.Players, event.ClientID)

			// send a message to connected clients to remove the player from the game
			playerDisconnect := &messages.ServerPlayerDisconnect{
				ClientID: event.ClientID,
			}
			payload, err := json.Marshal(playerDisconnect)
			if err != nil {
				log.Error("Failed to marshal player disconnect message: %v", err)
				continue
			}

			for _, client := range gm.clientManager.GetClients() {
				msg := &messages.Message{
					ClientID: 0, // ClientID 0 means the message is from the server
					Type:     messages.MessageTypeServerPlayerDisconnect,
					Payload:  payload,
				}

				err := network.WriteMessageToTCP(client.TCPConn, msg)
				if err != nil {
					log.Error("Failed to write message to TCP connection for client %d: %v", client.ID, err)
					continue
				}
			}
		default:
			log.Error("unhandled connection event type: %T", event)
		}
	}
}

// processClientMessages processes all pending client messages in the queue
// and updates the game state accordingly.
func (gm *GameManager) processClientMessages() {
	pendingMessages, err := gm.clientMessageQueue.ReadAllMessages()
	if err != nil {
		log.Error("Failed to read client messages: %v", err)
		return
	}
	for _, item := range pendingMessages {
		message, ok := item.(*messages.Message)
		if !ok {
			log.Error("Failed to cast message to messages.Message")
			continue
		}

		switch message.Type {
		case messages.MessageTypeClientPlayerUpdate:
			clientPlayerUpdate := &messages.ClientPlayerUpdate{}
			err := json.Unmarshal(message.Payload, clientPlayerUpdate)
			if err != nil {
				log.Error("Failed to unmarshal player state: %v", err)
				continue
			}
			if _, ok := gm.gameState.Players[message.ClientID]; !ok {
				log.Warn("Client %d is not in the game state", message.ClientID)
				continue
			}
			playerState := gm.gameState.Players[message.ClientID]

			if playerState.LastProcessedTimestamp > clientPlayerUpdate.Timestamp {
				log.Warn("Client %d sent an outdated player update", message.ClientID)
				continue
			}

			// TODO: check for previousUpdates that have not been processed
			// TODO: validate the update before applying it
			playerState.ApplyInput(clientPlayerUpdate)

			gm.checkPlayerCollisions(message.ClientID, playerState)
		default:
			log.Error("Unhandled message type: %s", message.Type)
		}
	}
}

// checkPlayerCollisions checks for collisions between a player and other objects in the game.
func (gm *GameManager) checkPlayerCollisions(clientID uint32, playerState *types.PlayerState) {
	// do attack hit detection
	if !playerState.IsAttackHitting {
		return
	}

	// create an attack hitbox for the player
	attackHitbox := playerState.Object.Clone()
	attackHitbox.Size.X = constants.PlayerAttackHitboxWidth
	if !playerState.AnimationFlip {
		attackHitbox.Position.X += constants.PlayerAttackHitboxOffset
	} else {
		attackHitbox.Position.X -= constants.PlayerAttackHitboxOffset
	}
	gm.gameState.CollisionSpace.Add(attackHitbox)

	// TODO: check for collision and get the ID from the collision shape data
	for npcID, npcState := range gm.gameState.NPCs {
		if !npcState.Exists() || npcState.IsDead() {
			continue
		}

		if !attackHitbox.SharesCells(npcState.Object) {
			continue
		}

		log.Debug("Player %d hit NPC %d", clientID, npcID)
		damage := constants.PlayerAttackDamage
		npcState.TakeDamage(damage)

		npcHit := &messages.ServerNPCHit{
			NPCID:    npcID,
			PlayerID: clientID,
			Damage:   damage,
		}
		payload, err := json.Marshal(npcHit)
		if err != nil {
			log.Error("Failed to marshal NPC hit message: %v", err)
			continue
		}

		for _, client := range gm.clientManager.GetClients() {
			msg := &messages.Message{
				ClientID: 0, // ClientID 0 means the message is from the server
				Type:     messages.MessageTypeServerNPCHit,
				Payload:  payload,
			}

			err := network.WriteMessageToTCP(client.TCPConn, msg)
			if err != nil {
				log.Error("Failed to write message to TCP connection for client %d: %v", client.ID, err)
				continue
			}
		}

		if !npcState.IsDead() {
			continue
		}

		log.Debug("Player %d killed NPC %d", clientID, npcID)
		npcKill := &messages.ServerNPCKill{
			NPCID:    npcID,
			PlayerID: clientID,
		}
		payload, err = json.Marshal(npcKill)
		if err != nil {
			log.Error("Failed to marshal NPC kill message: %v", err)
			continue
		}

		for _, client := range gm.clientManager.GetClients() {
			msg := &messages.Message{
				ClientID: 0, // ClientID 0 means the message is from the server
				Type:     messages.MessageTypeServerNPCKill,
				Payload:  payload,
			}

			err := network.WriteMessageToTCP(client.TCPConn, msg)
			if err != nil {
				log.Error("Failed to write message to TCP connection for client %d: %v", client.ID, err)
				continue
			}
		}
	}
}

// updateServerObjects updates server objects (e.g. npcs, items, projectiles, etc.)
func (gm *GameManager) updateServerObjects(deltaTime float64) {
	for _, npcState := range gm.gameState.NPCs {
		npcState.Update(deltaTime)
		if !npcState.Exists() {
			if npcState.RespawnTime() <= 0 {
				npcState.Spawn()
			} else {
				npcState.DecrementRespawnTime(deltaTime)
			}
		} else if npcState.IsDead() {
			npcState.Despawn()
		}
	}
}

// broadcastGameState sends the game state to connected clients.
func (gm *GameManager) broadcastGameState() {
	serverGameUpdate := ServerGameUpdateFromState(gm.gameState)
	payload, err := messages.SerializeGameState(serverGameUpdate)
	if err != nil {
		log.Error("Failed to serialize game state: %v", err)
		return
	}

	for _, client := range gm.clientManager.GetClients() {
		message := &messages.Message{
			ClientID: 0, // ClientID 0 means the message is from the server
			Type:     messages.MessageTypeServerGameUpdate,
			Payload:  payload,
		}

		if client.UDPAddress == nil {
			log.Trace("Client %d does not have a UDP address", client.ID)
			continue
		}

		err := network.WriteMessageToUDP(gm.clientManager.GetUDPConn(), client.UDPAddress, message)
		if err != nil {
			log.Error("Failed to write message to UDP connection for client %d: %v", client.ID, err)
			continue
		}
	}
}
