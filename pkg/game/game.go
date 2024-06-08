package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/workers"
	"github.com/solarlune/resolv"
)

type GameManager struct {
	gameState          *types.GameState
	clientMessageQueue queue.Queue
	serverEventQueue   queue.Queue
	repository         repositories.Repository
	saveStateChan      chan<- workers.SaveStateRequest
	serverMessageChan  chan<- workers.ServerMessage
	gameLoopInterval   time.Duration
	saveStateInterval  time.Duration
}

// NewGameManagerOptions contains options for creating a new GameManager.
type NewGameManagerOptions struct {
	ClientMessageQueue queue.Queue
	ServerEventQueue   queue.Queue
	Repository         repositories.Repository
	SaveStateChan      chan<- workers.SaveStateRequest
	ServerMessageChan  chan<- workers.ServerMessage
	GameLoopInterval   time.Duration
	SaveStateInterval  time.Duration
}

func NewGameManager(opts NewGameManagerOptions) *GameManager {
	return &GameManager{
		gameState:          types.NewGameState(NewCollisionSpace()),
		clientMessageQueue: opts.ClientMessageQueue,
		serverEventQueue:   opts.ServerEventQueue,
		repository:         opts.Repository,
		saveStateChan:      opts.SaveStateChan,
		serverMessageChan:  opts.ServerMessageChan,
		gameLoopInterval:   opts.GameLoopInterval,
		saveStateInterval:  opts.SaveStateInterval,
	}
}

// Start starts the game loop.
func (gm *GameManager) Start(ctx context.Context) error {
	if err := gm.initializeGameState(ctx); err != nil {
		return fmt.Errorf("failed to initialize game state: %v", err)
	}

	gameTicker := time.NewTicker(gm.gameLoopInterval)
	defer gameTicker.Stop()

	saveTicker := time.NewTicker(gm.saveStateInterval)
	defer saveTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-gameTicker.C:
			err := gm.gameTick(ctx, t)
			if err != nil {
				log.Error("Failed to run game tick: %v", err)
			}
			// TODO: server metrics
			// duration := time.Since(t)
			// log.Trace("Game tick took %s (%.2f%% of tick rate)", duration, float64(duration)/float64(gm.gameLoopInterval)*100)
		case <-saveTicker.C:
			saveRequest := workers.SaveStateRequest{
				Timestamp: gm.gameState.Timestamp,
				Type:      workers.SaveStateRequestTypeGame,
				State:     gm.gameState.Copy(),
			}
			gm.saveStateChan <- saveRequest
		}
	}
}

func (gm *GameManager) Stop() {
	// TODO: gracefully stop the game and save the game state
}

func (gm *GameManager) initializeGameState(_ context.Context) error {
	npcs := []struct {
		ID            uint32
		SpawnPosition kinematic.Vector
		Flip          bool
	}{
		{1, kinematic.NewVector(128-constants.NPCWidth/2, 16), false},
		{2, kinematic.NewVector(384-constants.NPCWidth/2, 16), false},
		{3, kinematic.NewVector(896-constants.NPCWidth/2, 16), true},
		{4, kinematic.NewVector(1152-constants.NPCWidth/2, 16), true},
	}

	for _, npc := range npcs {
		wanderRangeMinX := npc.SpawnPosition.X - constants.NPCWanderRange
		if wanderRangeMinX < 16 {
			wanderRangeMinX = 16
		}
		wanderRangeMaxX := npc.SpawnPosition.X + constants.NPCWanderRange
		if wanderRangeMaxX > float64(constants.SpaceWidth)-16-constants.NPCWidth {
			wanderRangeMaxX = float64(constants.SpaceWidth) - 16 - constants.NPCWidth
		}
		npcState := types.NewNPCState(npc.ID, npc.SpawnPosition, wanderRangeMinX, wanderRangeMaxX, npc.Flip)
		gm.gameState.NPCs[npc.ID] = npcState
		gm.gameState.CollisionSpace.Add(npcState.Object)
		npcState.Spawn()
	}

	return nil
}

// gameTick runs one iteration of the game loop.
func (gm *GameManager) gameTick(_ context.Context, t time.Time) error {
	gm.gameState.Timestamp = t.UnixMilli()
	gm.processServerEvents()
	gm.processClientMessages()
	gm.updateServerObjects(gm.gameLoopInterval.Seconds())
	gm.broadcastGameState()

	return nil
}

// processServerEvents processes all pending connection events in the queue,
// updates the game state, and notifies connected clients
func (gm *GameManager) processServerEvents() {
	pendingEvents, err := gm.serverEventQueue.ReadAllMessages()
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
	playerState := types.NewPlayerState(event.CharacterID, event.CharacterName, event.CharacterPosition, event.CharacterHitpoints)
	log.Debug("Client %d connected as %s", event.ClientID, event.CharacterName)
	// add the player to the game state
	gm.gameState.Players[event.ClientID] = playerState
	// add the player object to the collision space
	gm.gameState.CollisionSpace.Add(playerState.Object)

	playerConnect := &messages.ServerPlayerConnect{
		ClientID:    event.ClientID,
		PlayerState: PlayerStateUpdateFromState(playerState),
	}
	gm.serverMessageChan <- workers.ServerMessage{
		Type:    messages.MessageTypeServerPlayerConnect,
		Message: playerConnect,
	}

	return nil
}

func (gm *GameManager) handleDisconnectPlayerEvent(event *types.DisconnectPlayerEvent) error {
	// send a request to save the player state before deleting it
	saveRequest := workers.SaveStateRequest{
		Timestamp: gm.gameState.Timestamp,
		Type:      workers.SaveStateRequestTypePlayer,
		State:     gm.gameState.Players[event.ClientID].Copy(),
	}
	gm.saveStateChan <- saveRequest
	// remove the player object from the collision space
	gm.gameState.CollisionSpace.Remove(gm.gameState.Players[event.ClientID].Object)
	// npc follow target cleanup
	for _, npcState := range gm.gameState.NPCs {
		if npcState.FollowTarget == gm.gameState.Players[event.ClientID] {
			npcState.StopFollowing()
		}
	}
	// delete the player from the game state
	delete(gm.gameState.Players, event.ClientID)

	playerDisconnect := &messages.ServerPlayerDisconnect{
		ClientID: event.ClientID,
	}
	gm.serverMessageChan <- workers.ServerMessage{
		Type:    messages.MessageTypeServerPlayerDisconnect,
		Message: playerDisconnect,
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
	defer gm.gameState.CollisionSpace.Remove(attackHitbox)

	// TODO: check for collision and get the ID from the collision shape data
	for npcID, npcState := range gm.gameState.NPCs {
		if npcState.IsDead() {
			continue
		}

		if !attackHitbox.SharesCells(npcState.Object) {
			continue
		}

		// player hit npc
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
		gm.serverMessageChan <- workers.ServerMessage{
			Type:    messages.MessageTypeServerNPCHit,
			Message: npcHit,
		}

		if !npcState.IsDead() {
			if !npcState.IsFollowing() {
				npcState.StartFollowing(playerState)
			}
			continue
		}

		// player killed npc
		npcState.Despawn()

		log.Debug("Player %d killed NPC %d", clientID, npcID)
		npcKill := &messages.ServerNPCKill{
			NPCID:    npcID,
			PlayerID: clientID,
		}
		gm.serverMessageChan <- workers.ServerMessage{
			Type:    messages.MessageTypeServerNPCKill,
			Message: npcKill,
		}
	}
}

// updateServerObjects updates server objects (e.g. npcs, items, projectiles, etc.)
func (gm *GameManager) updateServerObjects(deltaTime float64) {
	for npcID, npcState := range gm.gameState.NPCs {
		if npcState.IsAttacking {
			// npc is attacking
			if npcState.IsAttackHitting {
				// npc is hitting with an attack
				gm.checkNPCAttackHit(npcID, npcState)
			}
		} else if !npcState.IsDead() {
			// npc is not following a player, so check for any players in line of sight
			gm.checkNPCLineOfSight(npcState)
		}

		npcState.Update(deltaTime)
	}
}

func (gm *GameManager) checkNPCAttackHit(npcID uint32, npcState *types.NPCState) {
	attackHitbox := npcState.Object.Clone()
	attackHitboxWidth, attackHitboxOffset := 0.0, 0.0
	switch npcState.CurrentAttack {
	case types.NPCAttack1:
		attackHitboxWidth = constants.NPCAttack1HitboxWidth
		attackHitboxOffset = constants.NPCAttack1HitboxOffset
	case types.NPCAttack2:
		attackHitboxWidth = constants.NPCAttack2HitboxWidth
		attackHitboxOffset = constants.NPCAttack2HitboxOffset
	case types.NPCAttack3:
		attackHitboxWidth = constants.NPCAttack3HitboxWidth
		attackHitboxOffset = constants.NPCAttack3HitboxOffset
	default:
		log.Warn("Unhandled NPC attack type: %d", npcState.CurrentAttack)
	}

	attackHitbox.Size.X = attackHitboxWidth
	if !npcState.AnimationFlip {
		attackHitbox.Position.X += attackHitboxOffset
	} else {
		attackHitbox.Position.X -= attackHitboxOffset
	}
	gm.gameState.CollisionSpace.Add(attackHitbox)
	defer gm.gameState.CollisionSpace.Remove(attackHitbox)

	for playerID, playerState := range gm.gameState.Players {
		if playerState.IsDead() {
			continue
		}

		if !attackHitbox.SharesCells(playerState.Object) {
			continue
		}

		log.Debug("NPC %d hit player %d", npcID, playerID)

		damage := int16(0)
		switch npcState.CurrentAttack {
		case types.NPCAttack1:
			damage = constants.NPCAttack1Damage
		case types.NPCAttack2:
			damage = constants.NPCAttack2Damage
		case types.NPCAttack3:
			damage = constants.NPCAttack3Damage
		default:
			log.Warn("Unhandled NPC attack type: %d", npcState.CurrentAttack)
		}

		playerState.TakeDamage(damage)

		playerHit := &messages.ServerPlayerHit{
			PlayerID: playerID,
			NPCID:    npcID,
			Damage:   damage,
		}
		gm.serverMessageChan <- workers.ServerMessage{
			Type:    messages.MessageTypeServerPlayerHit,
			Message: playerHit,
		}

		if !playerState.IsDead() {
			continue
		}

		log.Debug("NPC %d killed player %d", npcID, playerID)
		playerKill := &messages.ServerPlayerKill{
			PlayerID: playerID,
			NPCID:    npcID,
		}
		gm.serverMessageChan <- workers.ServerMessage{
			Type:    messages.MessageTypeServerPlayerKill,
			Message: playerKill,
		}
	}
}

func (gm *GameManager) checkNPCLineOfSight(npcState *types.NPCState) {
	flip := 1.0
	if npcState.AnimationFlip {
		flip = -1.0
	}
	lineOfSight := resolv.NewLine(npcState.Position.X+constants.NPCWidth/2, npcState.Position.Y+constants.NPCHeight/2, npcState.Position.X+constants.NPCWidth/2+flip*constants.NPCLineOfSight, npcState.Position.Y+constants.NPCHeight/2)
	for _, playerState := range gm.gameState.Players {
		if playerState.IsDead() {
			continue
		}
		if contact := lineOfSight.Intersection(0, 0, playerState.Object.Shape); contact != nil {
			npcState.StartFollowing(playerState)
			break
		}
	}
}

// broadcastGameState sends the game state to connected clients.
func (gm *GameManager) broadcastGameState() {
	// TODO: this is not scalable for large numbers of clients.
	// sending individual player and npc updates may be more efficient.
	// TODO: player vs localPlayer updates should be handled differently
	serverGameUpdate := ServerGameUpdateFromState(gm.gameState)
	gm.serverMessageChan <- workers.ServerMessage{
		Type:    messages.MessageTypeServerGameUpdate,
		Message: serverGameUpdate,
	}
}
