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
			// TODO: server metrics
			// duration := time.Since(t)
			// log.Trace("Game tick took %s (%.2f%% of tick rate)", duration, float64(duration)/float64(gm.gameLoopInterval)*100)
		}
	}
}

func (gm *GameManager) Stop() {
	// TODO: gracefully stop the game and save the game state
}

func (gm *GameManager) initializeGameState(_ context.Context) error {
	npc1State := types.NewNPCState(128.0-constants.NPCWidth/2, 16.0)
	gm.gameState.NPCs[1] = npc1State
	gm.gameState.CollisionSpace.Add(npc1State.Object)

	npc2State := types.NewNPCState(384.0-constants.NPCWidth/2, 16.0)
	gm.gameState.NPCs[2] = npc2State
	gm.gameState.CollisionSpace.Add(npc2State.Object)

	npc3State := types.NewNPCState(896.0-constants.NPCWidth/2, 16.0)
	npc3State.AnimationFlip = true
	gm.gameState.NPCs[3] = npc3State
	gm.gameState.CollisionSpace.Add(npc3State.Object)

	npc4State := types.NewNPCState(1152.0-constants.NPCWidth/2, 16.0)
	npc4State.AnimationFlip = true
	gm.gameState.NPCs[4] = npc4State
	gm.gameState.CollisionSpace.Add(npc4State.Object)

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
			if err := gm.handleConnectPlayerEvent(event); err != nil {
				log.Error("Failed to handle connect player event: %v", err)
			}
		case *types.DisconnectPlayerEvent:
			if err := gm.handleDisconnectPlayerEvent(event); err != nil {
				log.Error("Failed to handle disconnect player event: %v", err)
			}
		default:
			log.Error("unhandled connection event type: %T", event)
		}
	}
}

func (gm *GameManager) handleConnectPlayerEvent(event *types.ConnectPlayerEvent) error {
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
		return fmt.Errorf("failed to marshal player state: %v", err)
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

	return nil
}

func (gm *GameManager) handleDisconnectPlayerEvent(event *types.DisconnectPlayerEvent) error {
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
		return fmt.Errorf("failed to marshal player disconnect message: %v", err)
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

	return nil
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
			// TODO: investigate spikes in messages from a given client
			if err := gm.handleClientPlayerUpdate(message); err != nil {
				log.Error("Failed to handle client player update: %v", err)
			}
		default:
			log.Error("Unhandled message type: %s", message.Type)
		}
	}
}

func (gm *GameManager) handleClientPlayerUpdate(message *messages.Message) error {
	clientPlayerUpdate := &messages.ClientPlayerUpdate{}
	err := json.Unmarshal(message.Payload, clientPlayerUpdate)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client player update: %v", err)
	}
	if _, ok := gm.gameState.Players[message.ClientID]; !ok {
		log.Warn("Client %d is not in the game state", message.ClientID)
		return nil
	}
	playerState := gm.gameState.Players[message.ClientID]

	if playerState.LastProcessedTimestamp > clientPlayerUpdate.Timestamp {
		log.Warn("Client %d sent an outdated player update", message.ClientID)
		return nil
	}

	// check for past updates that have not been processed and apply first
	for _, pastUpdate := range clientPlayerUpdate.PastUpdates {
		if playerState.LastProcessedTimestamp >= pastUpdate.Timestamp {
			continue
		}
		log.Warn("Applying past update for client %d - Last processed: %d, past update: %d", message.ClientID, playerState.LastProcessedTimestamp, pastUpdate.Timestamp)
		playerState.ApplyInput(pastUpdate)
	}

	// TODO: validate the update before applying it
	playerState.ApplyInput(clientPlayerUpdate)

	gm.checkPlayerCollisions(message.ClientID, playerState)

	return nil
}

// checkPlayerCollisions checks for collisions between a player and other objects in the game.
func (gm *GameManager) checkPlayerCollisions(clientID uint32, playerState *types.PlayerState) {
	// do attack hit detection
	if !playerState.IsAttackHitting {
		return
	}

	// create an attack hitbox for the player
	attackHitbox := playerState.Object.Clone()

	attackHitboxWidth, attackHitboxOffset := 0.0, 0.0
	switch playerState.CurrentAttack {
	case types.PlayerAttack1:
		attackHitboxWidth = constants.PlayerAttack1HitboxWidth
		attackHitboxOffset = constants.PlayerAttack1HitboxOffset
	case types.PlayerAttack2:
		attackHitboxWidth = constants.PlayerAttack2HitboxWidth
		attackHitboxOffset = constants.PlayerAttack2HitboxOffset
	case types.PlayerAttack3:
		attackHitboxWidth = constants.PlayerAttack3HitboxWidth
		attackHitboxOffset = constants.PlayerAttack3HitboxOffset
	default:
		log.Warn("Unhandled player attack type: %d", playerState.CurrentAttack)
	}

	attackHitbox.Size.X = attackHitboxWidth
	if !playerState.AnimationFlip {
		attackHitbox.Position.X += attackHitboxOffset
	} else {
		attackHitbox.Position.X -= attackHitboxOffset
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

		damage := int16(0)
		switch playerState.CurrentAttack {
		case types.PlayerAttack1:
			damage = constants.PlayerAttack1Damage
		case types.PlayerAttack2:
			damage = constants.PlayerAttack2Damage
		case types.PlayerAttack3:
			damage = constants.PlayerAttack3Damage
		default:
			log.Warn("Unhandled player attack type: %d", playerState.CurrentAttack)
		}

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
	// TODO: this is not scalable for large numbers of clients.
	// sending individual player and npc updates may be more efficient.

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
