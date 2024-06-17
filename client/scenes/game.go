package scenes

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
)

const (
	// Zoom is the zoom scale of the game viewport.
	Zoom = 1.0
	// InterpolationOffset is how far back in time we want to interpolate.
	InterpolationOffset = 150 // ms - currently 3x the server tick rate (50ms)
)

type GameScene struct {
	*BaseScene

	// networkManager is the network manager.
	networkManager *network.NetworkManager
	// collisionSpace is the collision space.
	collisionSpace *resolv.Space
	// world is the world image.
	world *ebiten.Image
	// CameraViewport is the current viewport.
	CameraViewport *CameraViewport
	// deletedObjects is a map of deleted game objects indexed by a unique identifier
	// and the timestamp of the deletion.
	deletedObjects map[string]int64

	serverPlayerUpdateBuffers map[uint32]*ServerPlayerUpdateBuffer
	serverNPCUpdateBuffers    map[uint32]*ServerNPCUpdateBuffer
}

type CameraViewport struct {
	X int
	Y int
}

type ServerPlayerUpdateBuffer struct {
	LastUpdateReceived int64
	Updates            []*messages.ServerPlayerUpdate
}

type ServerNPCUpdateBuffer struct {
	LastStateReceived int64
	States            []*messages.ServerNPCUpdate
}

var _ Scene = &GameScene{}

func NewGameScene(networkManager *network.NetworkManager) (Scene, error) {
	collisionSpace := game.NewCollisionSpace()
	world := ebiten.NewImage(collisionSpace.Width()*collisionSpace.CellWidth, collisionSpace.Height()*collisionSpace.CellHeight)
	return &GameScene{
		BaseScene:                 NewBaseScene(objects.NewSortedZIndexObject("game-root")),
		networkManager:            networkManager,
		collisionSpace:            game.NewCollisionSpace(),
		world:                     world,
		deletedObjects:            make(map[string]int64),
		serverPlayerUpdateBuffers: make(map[uint32]*ServerPlayerUpdateBuffer),
		serverNPCUpdateBuffers:    make(map[uint32]*ServerNPCUpdateBuffer),
	}, nil
}

func (g *GameScene) Update() error {
	if err := g.processPendingServerMessages(); err != nil {
		return fmt.Errorf("failed to process pending server messages: %v", err)
	}

	if err := g.updateObjectStates(); err != nil {
		return fmt.Errorf("failed to update object states: %v", err)
	}

	if err := g.BaseScene.Update(); err != nil {
		return fmt.Errorf("failed to update base scene: %v", err)
	}

	if err := g.cleanupDeletedObjects(); err != nil {
		return fmt.Errorf("failed to cleanup deleted objects: %v", err)
	}

	return nil
}

func (g *GameScene) processPendingServerMessages() error {
	serverMessages, err := g.networkManager.ServerMessageQueue().ReadAllMessages()
	if err != nil {
		return fmt.Errorf("failed to read server messages: %v", err)
	}

	for _, item := range serverMessages {
		message, ok := item.(*messages.Message)
		if !ok {
			log.Error("Failed to cast message to messages.Message")
			continue
		}

		switch message.Type {
		case messages.MessageTypeServerPlayerUpdate:
			if err := g.handleServerPlayerUpdate(message); err != nil {
				log.Error("Failed to handle server player update: %v", err)
			}
		case messages.MessageTypeServerNPCUpdate:
			if err := g.handleServerNPCUpdate(message); err != nil {
				log.Error("Failed to handle server NPC update: %v", err)
			}
		case messages.MessageTypeServerPlayerConnect:
			if err := g.handleServerPlayerConnect(message); err != nil {
				log.Error("Failed to handle server player connect: %v", err)
			}
		case messages.MessageTypeServerPlayerDisconnect:
			if err := g.handleServerPlayerDisconnect(message); err != nil {
				log.Error("Failed to handle server player disconnect: %v", err)
			}
		case messages.MessageTypeServerNPCHit:
			if err := g.handleServerNPCHit(message); err != nil {
				log.Error("Failed to handle server NPC hit: %v", err)
			}
		case messages.MessageTypeServerNPCKill:
			if err := g.handleServerNPCKill(message); err != nil {
				log.Error("Failed to handle server NPC kill: %v", err)
			}
		case messages.MessageTypeServerPlayerHit:
			if err := g.handleServerPlayerHit(message); err != nil {
				log.Error("Failed to handle server player hit: %v", err)
			}
		case messages.MessageTypeServerPlayerKill:
			if err := g.handleServerPlayerKill(message); err != nil {
				log.Error("Failed to handle server player kill: %v", err)
			}
		default:
			log.Warn("Received unexpected message type from server: %s", message.Type)
		}
	}

	return nil
}

func (g *GameScene) getLocalPlayer() (*objects.Player, error) {
	if !g.networkManager.IsConnected() {
		return nil, nil
	}
	id := fmt.Sprintf("player-%d", g.networkManager.ClientID())
	obj := g.GetRoot().GetChild(id)
	if obj == nil {
		return nil, nil
	}
	playerObject, ok := obj.(*objects.Player)
	if !ok {
		return nil, fmt.Errorf("failed to cast game object %s to *objects.Player", id)
	}
	return playerObject, nil
}

func (g *GameScene) handleServerPlayerUpdate(message *messages.Message) error {
	serverPlayerUpdate, err := messages.DeserializeServerPlayerUpdate(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to deserialize player update: %v", err)
	}

	if _, ok := g.serverPlayerUpdateBuffers[serverPlayerUpdate.ClientID]; !ok {
		g.serverPlayerUpdateBuffers[serverPlayerUpdate.ClientID] = &ServerPlayerUpdateBuffer{
			LastUpdateReceived: serverPlayerUpdate.Timestamp,
			Updates:            []*messages.ServerPlayerUpdate{serverPlayerUpdate},
		}
		return nil
	}

	serverPlayerUpdateBuffer := g.serverPlayerUpdateBuffers[serverPlayerUpdate.ClientID]
	if serverPlayerUpdate.Timestamp < serverPlayerUpdateBuffer.LastUpdateReceived {
		log.Warn("Received outdated player state: %d < %d", serverPlayerUpdate.Timestamp, serverPlayerUpdateBuffer.LastUpdateReceived)
		return nil
	}
	serverPlayerUpdateBuffer.LastUpdateReceived = serverPlayerUpdate.Timestamp
	serverPlayerUpdateBuffer.Updates = append(serverPlayerUpdateBuffer.Updates, serverPlayerUpdate)

	if serverPlayerUpdate.ClientID != g.networkManager.ClientID() {
		return nil
	}

	localPlayer, err := g.getLocalPlayer()
	if err != nil {
		return fmt.Errorf("failed to get local player: %v", err)
	}
	if localPlayer == nil {
		return nil
	}

	playerState := game.PlayerStateFromServerUpdate(serverPlayerUpdate.PlayerState)
	if err := localPlayer.ReconcileState(playerState); err != nil {
		return fmt.Errorf("failed to reconcile local player state: %v", err)
	}

	return nil
}

func (g *GameScene) handleServerNPCUpdate(message *messages.Message) error {
	serverNPCUpdate, err := messages.DeserializeServerNPCUpdate(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to deserialize NPC update: %v", err)
	}

	if _, ok := g.serverNPCUpdateBuffers[serverNPCUpdate.NPCID]; !ok {
		g.serverNPCUpdateBuffers[serverNPCUpdate.NPCID] = &ServerNPCUpdateBuffer{
			LastStateReceived: serverNPCUpdate.Timestamp,
			States:            []*messages.ServerNPCUpdate{serverNPCUpdate},
		}
		return nil
	}

	serverNPCUpdateBuffer := g.serverNPCUpdateBuffers[serverNPCUpdate.NPCID]
	if serverNPCUpdate.Timestamp < serverNPCUpdateBuffer.LastStateReceived {
		log.Warn("Received outdated NPC state: %d < %d", serverNPCUpdate.Timestamp, serverNPCUpdateBuffer.LastStateReceived)
		return nil
	}
	serverNPCUpdateBuffer.LastStateReceived = serverNPCUpdate.Timestamp
	serverNPCUpdateBuffer.States = append(serverNPCUpdateBuffer.States, serverNPCUpdate)

	return nil
}

func (g *GameScene) handleServerPlayerConnect(message *messages.Message) error {
	playerConnect := &messages.ServerPlayerConnect{}
	if err := json.Unmarshal(message.Payload, playerConnect); err != nil {
		return fmt.Errorf("failed to unmarshal player connect message: %v", err)
	}

	id := fmt.Sprintf("player-%d", playerConnect.ClientID)
	obj := g.GetRoot().GetChild(id)
	if obj != nil {
		log.Warn("Player object for client %d already exists", playerConnect.ClientID)
		return nil
	}
	log.Debug("Adding new player object for client %d", playerConnect.ClientID)
	playerState := game.PlayerStateFromServerUpdate(playerConnect.PlayerState)
	playerObject, err := objects.NewPlayer(id, g.networkManager, playerState)
	if err != nil {
		return fmt.Errorf("failed to create new player object: %v", err)
	}
	g.collisionSpace.Add(playerObject.State.Object)
	if err := g.GetRoot().AddChild(id, playerObject); err != nil {
		return fmt.Errorf("failed to add player object: %v", err)
	}
	delete(g.deletedObjects, id)

	return nil
}

func (g *GameScene) handleServerPlayerDisconnect(message *messages.Message) error {
	playerDisconnect := &messages.ServerPlayerDisconnect{}
	if err := json.Unmarshal(message.Payload, playerDisconnect); err != nil {
		return fmt.Errorf("failed to unmarshal player disconnect message: %v", err)
	}

	id := fmt.Sprintf("player-%d", playerDisconnect.ClientID)
	obj := g.GetRoot().GetChild(id)
	if obj == nil {
		log.Warn("Player object for client %d not found", playerDisconnect.ClientID)
		return nil
	}
	playerObject, ok := obj.(*objects.Player)
	if !ok {
		return fmt.Errorf("failed to cast game object %s to *objects.Player", id)
	}
	log.Debug("Removing player object for client %d", playerDisconnect.ClientID)
	g.collisionSpace.Remove(playerObject.State.Object)
	if err := g.GetRoot().RemoveChild(id); err != nil {
		return fmt.Errorf("failed to remove player object: %v", err)
	}
	g.deletedObjects[id] = time.Now().UnixMilli()

	// remove the player from the server update buffer
	delete(g.serverPlayerUpdateBuffers, playerDisconnect.ClientID)

	return nil
}

func (g *GameScene) handleServerNPCHit(message *messages.Message) error {
	npcHit := &messages.ServerNPCHit{}
	if err := json.Unmarshal(message.Payload, npcHit); err != nil {
		return fmt.Errorf("failed to unmarshal NPC hit message: %v", err)
	}
	log.Debug("Player %d hit NPC %d for %d damage", npcHit.PlayerID, npcHit.NPCID, npcHit.Damage)
	npcID := fmt.Sprintf("npc-%d", npcHit.NPCID)
	obj := g.GetRoot().GetChild(npcID)
	if obj == nil {
		log.Warn("NPC object with id %d not found", npcHit.NPCID)
		return nil
	}
	npcObject, ok := obj.(*objects.NPC)
	if !ok {
		return fmt.Errorf("failed to cast game object %s to *objects.NPC", npcID)
	}

	hitID := fmt.Sprintf("%s-hit-%d", npcObject.ID, uuid.New().ID())
	zIndex := 15
	if npcHit.PlayerID == g.networkManager.ClientID() {
		zIndex = 25
	}
	hitObject := objects.NewTextEffect(hitID, objects.NewTextEffectOptions{
		Text:   fmt.Sprintf("%d", npcHit.Damage),
		X:      npcObject.State.Position.X + constants.NPCWidth/2,
		Y:      npcObject.State.Position.Y + constants.NPCHeight/2,
		Color:  color.RGBA{255, 0, 0, 255}, // Red
		Scroll: true,
		TTL:    1500,
		ZIndex: zIndex,
	})
	if err := g.GetRoot().AddChild(hitID, hitObject); err != nil {
		return fmt.Errorf("failed to add text effect: %v", err)
	}

	return nil
}

func (g *GameScene) handleServerNPCKill(message *messages.Message) error {
	npcKill := &messages.ServerNPCKill{}
	if err := json.Unmarshal(message.Payload, npcKill); err != nil {
		return fmt.Errorf("failed to unmarshal NPC kill message: %v", err)
	}
	log.Debug("Player %d killed NPC %d", npcKill.PlayerID, npcKill.NPCID)
	return nil
}

func (g *GameScene) handleServerPlayerHit(message *messages.Message) error {
	playerHit := &messages.ServerPlayerHit{}
	if err := json.Unmarshal(message.Payload, playerHit); err != nil {
		return fmt.Errorf("failed to unmarshal player hit message: %v", err)
	}
	log.Debug("NPC %d hit player %d for %d damage", playerHit.NPCID, playerHit.PlayerID, playerHit.Damage)
	playerID := fmt.Sprintf("player-%d", playerHit.PlayerID)
	obj := g.GetRoot().GetChild(playerID)
	if obj == nil {
		log.Warn("Player object with id %d not found", playerHit.PlayerID)
		return nil
	}
	playerObject, ok := obj.(*objects.Player)
	if !ok {
		return fmt.Errorf("failed to cast game object %s to *objects.Player", playerID)
	}

	zIndex := 25
	isLocalPlayer := playerHit.PlayerID == g.networkManager.ClientID()
	if isLocalPlayer {
		zIndex = 35
	}

	hitID := fmt.Sprintf("%s-hit-%d", playerObject.ID, uuid.New().ID())
	hitObject := objects.NewTextEffect(hitID, objects.NewTextEffectOptions{
		Text:   fmt.Sprintf("%d", playerHit.Damage),
		X:      playerObject.State.Position.X + constants.PlayerWidth/2,
		Y:      playerObject.State.Position.Y + constants.PlayerHeight/2,
		Color:  color.RGBA{255, 0, 0, 255}, // Red
		Scroll: true,
		TTL:    1500,
		ZIndex: zIndex,
	})
	if err := g.GetRoot().AddChild(hitID, hitObject); err != nil {
		return fmt.Errorf("failed to add text effect: %v", err)
	}

	return nil
}

func (g *GameScene) handleServerPlayerKill(message *messages.Message) error {
	playerKill := &messages.ServerPlayerKill{}
	if err := json.Unmarshal(message.Payload, playerKill); err != nil {
		return fmt.Errorf("failed to unmarshal player kill message: %v", err)
	}
	log.Debug("NPC %d killed player %d", playerKill.NPCID, playerKill.PlayerID)
	return nil
}

func (g *GameScene) updateObjectStates() error {
	serverTime, _ := g.networkManager.ServerTime()
	renderTime := int64(math.Round(serverTime)) - InterpolationOffset

	for clientID, serverPlayerUpdateBuffer := range g.serverPlayerUpdateBuffers {
		if len(serverPlayerUpdateBuffer.Updates) < 2 {
			continue
		}

		for len(serverPlayerUpdateBuffer.Updates) > 2 && serverPlayerUpdateBuffer.Updates[2].Timestamp < renderTime {
			serverPlayerUpdateBuffer.Updates = serverPlayerUpdateBuffer.Updates[1:]
		}

		if len(serverPlayerUpdateBuffer.Updates) > 2 {
			// we have a future state to interpolate to
			from := serverPlayerUpdateBuffer.Updates[1]
			to := serverPlayerUpdateBuffer.Updates[2]
			interpolationFactor := float64(renderTime-from.Timestamp) / float64(to.Timestamp-from.Timestamp)
			fromState := game.PlayerStateFromServerUpdate(from.PlayerState)
			toState := game.PlayerStateFromServerUpdate(to.PlayerState)
			if err := g.interpolatePlayerState(clientID, fromState, toState, interpolationFactor); err != nil {
				log.Warn("Failed to interpolate player state: %v", err)
				continue
			}
		} else {
			// we don't have a future state, so we need to extrapolate from the last state
			from := serverPlayerUpdateBuffer.Updates[0]
			to := serverPlayerUpdateBuffer.Updates[1]
			extrapolationFactor := float64(renderTime-to.Timestamp) / float64(to.Timestamp-from.Timestamp)
			fromState := game.PlayerStateFromServerUpdate(from.PlayerState)
			toState := game.PlayerStateFromServerUpdate(to.PlayerState)
			if err := g.extrapolatePlayerState(clientID, fromState, toState, extrapolationFactor); err != nil {
				log.Warn("Failed to extrapolate player state: %v", err)
				continue
			}
		}
	}

	for npcID, serverNPCUpdateBuffer := range g.serverNPCUpdateBuffers {
		if len(serverNPCUpdateBuffer.States) < 2 {
			continue
		}

		for len(serverNPCUpdateBuffer.States) > 2 && serverNPCUpdateBuffer.States[2].Timestamp < renderTime {
			serverNPCUpdateBuffer.States = serverNPCUpdateBuffer.States[1:]
		}

		if len(serverNPCUpdateBuffer.States) > 2 {
			// we have a future state to interpolate to
			from := serverNPCUpdateBuffer.States[1]
			to := serverNPCUpdateBuffer.States[2]
			interpolationFactor := float64(renderTime-from.Timestamp) / float64(to.Timestamp-from.Timestamp)
			fromState := game.NPCStateFromServerUpdate(from.NPCState)
			toState := game.NPCStateFromServerUpdate(to.NPCState)
			if err := g.interpolateNPCState(npcID, fromState, toState, interpolationFactor); err != nil {
				log.Warn("Failed to interpolate NPC state: %v", err)
				continue
			}
		} else {
			// we don't have a future state, so we need to extrapolate from the last state
			from := serverNPCUpdateBuffer.States[0]
			to := serverNPCUpdateBuffer.States[1]
			extrapolationFactor := float64(renderTime-to.Timestamp) / float64(to.Timestamp-from.Timestamp)
			fromState := game.NPCStateFromServerUpdate(from.NPCState)
			toState := game.NPCStateFromServerUpdate(to.NPCState)
			if err := g.extrapolateNPCState(npcID, fromState, toState, extrapolationFactor); err != nil {
				log.Warn("Failed to extrapolate NPC state: %v", err)
				continue
			}
		}
	}

	return nil
}

func (g *GameScene) interpolatePlayerState(clientID uint32, from *gametypes.PlayerState, to *gametypes.PlayerState, factor float64) error {
	id := fmt.Sprintf("player-%d", clientID)
	obj := g.GetRoot().GetChild(id)
	if obj == nil {
		// TODO: handle edge case where the client misses the disconnect message, but receives some updates with the player still in the game
		if _, ok := g.deletedObjects[id]; ok {
			log.Debug("Player object for client %d was recently deleted, not instancing as part of update", clientID)
			return nil
		}
		log.Debug("Adding new player object for client %d", clientID)
		playerObject, err := objects.NewPlayer(id, g.networkManager, to)
		if err != nil {
			return fmt.Errorf("failed to create new player object: %v", err)
		}
		g.collisionSpace.Add(playerObject.State.Object)
		if err := g.GetRoot().AddChild(id, playerObject); err != nil {
			return fmt.Errorf("failed to add player object: %v", err)
		}
	} else {
		if clientID == g.networkManager.ClientID() {
			return nil
		}
		playerObject, ok := obj.(*objects.Player)
		if !ok {
			return fmt.Errorf("failed to cast game object %s to *objects.Player", id)
		}
		playerObject.InterpolateState(from, to, factor)
	}

	return nil
}

func (g *GameScene) interpolateNPCState(npcID uint32, from *gametypes.NPCState, to *gametypes.NPCState, factor float64) error {
	id := fmt.Sprintf("npc-%d", npcID)
	obj := g.GetRoot().GetChild(id)
	if obj == nil {
		// TODO: handle edge case where the client misses the despawn message, but receives some updates with the npc still in the game
		if _, ok := g.deletedObjects[id]; ok {
			log.Debug("NPC object with id %d was recently deleted, not instancing as part of update", npcID)
			return nil
		}
		log.Debug("Adding new npc object with id %d", npcID)
		npcObject, err := objects.NewNPC(id, to)
		if err != nil {
			return fmt.Errorf("failed to create new player object: %v", err)
		}
		// we don't need to add NPCs to the client's collision space
		if err := g.GetRoot().AddChild(id, npcObject); err != nil {
			return fmt.Errorf("failed to add npc object: %v", err)
		}
	} else {
		npcObject, ok := obj.(*objects.NPC)
		if !ok {
			return fmt.Errorf("failed to cast game object %s to *objects.NPC", id)
		}
		npcObject.InterpolateState(from, to, factor)
	}

	return nil
}

// extrapolateState extrapolates the game state based on the last two states.
// This is used when we don't have a future state to interpolate to.
func (g *GameScene) extrapolatePlayerState(clientID uint32, from *gametypes.PlayerState, to *gametypes.PlayerState, factor float64) error {
	id := fmt.Sprintf("player-%d", clientID)
	obj := g.GetRoot().GetChild(id)
	if obj == nil {
		log.Debug("Player object for client %d not found, not instancing since we're extrapolating", clientID)
	} else {
		if clientID == g.networkManager.ClientID() {
			return nil
		}
		playerObject, ok := obj.(*objects.Player)
		if !ok {
			return fmt.Errorf("failed to cast game object %s to *objects.Player", id)
		}
		playerObject.ExtrapolateState(from, to, factor)
	}

	return nil
}

func (g *GameScene) extrapolateNPCState(npcID uint32, from *gametypes.NPCState, to *gametypes.NPCState, factor float64) error {
	id := fmt.Sprintf("npc-%d", npcID)
	obj := g.GetRoot().GetChild(id)
	if obj == nil {
		log.Debug("NPC object with id %d not found, not instancing since we're extrapolating", npcID)
	} else {
		npcObject, ok := obj.(*objects.NPC)
		if !ok {
			return fmt.Errorf("failed to cast game object %s to *objects.NPC", id)
		}
		npcObject.ExtrapolateState(from, to, factor)
	}

	return nil
}

func (g *GameScene) cleanupDeletedObjects() error {
	for id, timestamp := range g.deletedObjects {
		if time.Now().UnixMilli()-timestamp > 2000 {
			delete(g.deletedObjects, id)
		}
	}
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	localPlayer, err := g.getLocalPlayer()
	if err != nil {
		log.Error("Failed to get local player: %v", err)
		return
	}
	if localPlayer == nil {
		log.Debug("Not drawing game scene because local player object not found")
		return
	}
	g.drawWorld()
	g.drawViewport(screen, localPlayer, Zoom)
}

func (g *GameScene) drawWorld() {
	// Draw the world background
	vector.DrawFilledRect(g.world, 0, 0, float32(g.world.Bounds().Dx()), float32(g.world.Bounds().Dy()), color.RGBA{0x87, 0xce, 0xeb, 0xff}, false)

	for _, obj := range g.collisionSpace.Objects() {
		if obj.HasTags(gametypes.CollisionSpaceTagLevel) {
			levelColor := color.RGBA{0x80, 0x80, 0x80, 0xff} // Gray
			vector.DrawFilledRect(g.world, float32(obj.Position.X), float32(g.world.Bounds().Dy())-float32(obj.Position.Y)-float32(obj.Size.Y), float32(obj.Size.X), float32(obj.Size.Y), levelColor, false)
		} else if obj.HasTags(gametypes.CollisionSpaceTagPlatform) {
			platformColor := color.RGBA{0xff, 0xa5, 0x00, 0xff} // Orange
			vector.DrawFilledRect(g.world, float32(obj.Position.X), float32(g.world.Bounds().Dy())-float32(obj.Position.Y)-float32(obj.Size.Y), float32(obj.Size.X), float32(obj.Size.Y), platformColor, false)
		} else if obj.HasTags(gametypes.CollisionSpaceTagLadder) {
			ladderColor := color.RGBA{0x00, 0x80, 0x00, 0xff} // Green
			vector.DrawFilledRect(g.world, float32(obj.Position.X), float32(g.world.Bounds().Dy())-float32(obj.Position.Y)-float32(obj.Size.Y), float32(obj.Size.X), float32(obj.Size.Y), ladderColor, false)
		}
	}

	g.BaseScene.Draw(g.world)
}

func (g *GameScene) drawViewport(screen *ebiten.Image, player *objects.Player, zoom float64) {
	// calculate the viewport center based on the player position and the viewport center
	vx, vy := int(player.State.Position.X+constants.PlayerWidth/2), g.world.Bounds().Dy()-int(player.State.Position.Y)-int(constants.PlayerHeight/2)
	g.CameraViewport = &CameraViewport{X: vx, Y: vy}
	// TODO: smooth the camera movement

	// calculate the viewport bounds based on the zoom level
	zoomFactor := 1.0 / (zoom * 2)
	minX, maxX := float64(g.CameraViewport.X)-float64(screen.Bounds().Dx())*zoomFactor, float64(g.CameraViewport.X)+float64(screen.Bounds().Dx())*zoomFactor
	minY, maxY := float64(g.CameraViewport.Y)-float64(screen.Bounds().Dy())*zoomFactor, float64(g.CameraViewport.Y)+float64(screen.Bounds().Dy())*zoomFactor

	// clamp the viewport to the world bounds
	if minX < 0 {
		maxX += math.Abs(minX)
		minX = 0
	}
	if maxX > float64(g.world.Bounds().Dx()) {
		minX -= maxX - float64(g.world.Bounds().Dx())
		maxX = float64(g.world.Bounds().Dx())
	}
	if minY < 0 {
		maxY += math.Abs(minY)
		minY = 0
	}
	if maxY > float64(g.world.Bounds().Dy()) {
		minY -= maxY - float64(g.world.Bounds().Dy())
		maxY = float64(g.world.Bounds().Dy())
	}

	viewport := g.world.SubImage(image.Rectangle{
		Min: image.Point{X: int(minX), Y: int(minY)},
		Max: image.Point{X: int(maxX), Y: int(maxY)},
	}).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(zoom, zoom)

	screen.DrawImage(viewport, opts)
}
