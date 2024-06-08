package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/network"
)

type BroadcastMessageWorker struct {
	clientManager        *network.ClientManager
	broadcastMessageChan <-chan BroadcastMessage
}

type BroadcastMessage struct {
	Type    messages.MessageType
	Message interface{}
}

type NewBroadcastMessageWorkerOptions struct {
	ClientManager        *network.ClientManager
	BroadcastMessageChan <-chan BroadcastMessage
}

func NewBroadcastMessageWorker(opts NewBroadcastMessageWorkerOptions) *BroadcastMessageWorker {
	return &BroadcastMessageWorker{
		clientManager:        opts.ClientManager,
		broadcastMessageChan: opts.BroadcastMessageChan,
	}
}

func (w *BroadcastMessageWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-w.broadcastMessageChan:
			switch msg.Type {
			case messages.MessageTypeServerPlayerConnect:
				if err := w.handleServerPlayerConnect(msg); err != nil {
					log.Error("Failed to handle server player connect message: %v", err)
				}
			case messages.MessageTypeServerPlayerDisconnect:
				if err := w.handleServerPlayerDisconnect(msg); err != nil {
					log.Error("Failed to handle server player disconnect message: %v", err)
				}
			case messages.MessageTypeServerGameUpdate:
				if err := w.handleServerGameUpdate(msg); err != nil {
					log.Error("Failed to handle server game update message: %v", err)
				}
			case messages.MessageTypeServerPlayerUpdate:
				if err := w.handleServerPlayerUpdate(msg); err != nil {
					log.Error("Failed to handle server player update message: %v", err)
				}
			case messages.MessageTypeServerNPCUpdate:
				if err := w.handleServerNPCUpdate(msg); err != nil {
					log.Error("Failed to handle server NPC update message: %v", err)
				}
			case messages.MessageTypeServerNPCHit:
				if err := w.handleServerNPCHit(msg); err != nil {
					log.Error("Failed to handle server NPC hit message: %v", err)
				}
			case messages.MessageTypeServerNPCKill:
				if err := w.handleServerNPCKill(msg); err != nil {
					log.Error("Failed to handle server NPC kill message: %v", err)
				}
			case messages.MessageTypeServerPlayerHit:
				if err := w.handleServerPlayerHit(msg); err != nil {
					log.Error("Failed to handle server player hit message: %v", err)
				}
			case messages.MessageTypeServerPlayerKill:
				if err := w.handleServerPlayerKill(msg); err != nil {
					log.Error("Failed to handle server player kill message: %v", err)
				}
			default:
				log.Error("Unknown server message type: %v", msg.Type)
			}
		}
	}
}

func (w *BroadcastMessageWorker) handleServerPlayerConnect(msg BroadcastMessage) error {
	playerConnect, ok := msg.Message.(*messages.ServerPlayerConnect)
	if !ok {
		return fmt.Errorf("failed to cast server player connect message")
	}

	payload, err := json.Marshal(playerConnect)
	if err != nil {
		return fmt.Errorf("failed to marshal player state: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		msg := &messages.Message{
			ClientID: 0,
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

func (w *BroadcastMessageWorker) handleServerPlayerDisconnect(msg BroadcastMessage) error {
	playerDisconnect, ok := msg.Message.(*messages.ServerPlayerDisconnect)
	if !ok {
		return fmt.Errorf("failed to cast server player disconnect message")
	}

	payload, err := json.Marshal(playerDisconnect)
	if err != nil {
		return fmt.Errorf("failed to marshal player disconnect message: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		msg := &messages.Message{
			ClientID: 0,
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

func (w *BroadcastMessageWorker) handleServerGameUpdate(msg BroadcastMessage) error {
	serverGameUpdate, ok := msg.Message.(*messages.ServerGameUpdate)
	if !ok {
		return fmt.Errorf("failed to cast server game update message")
	}

	payload, err := messages.SerializeGameState(serverGameUpdate)
	if err != nil {
		return fmt.Errorf("failed to serialize game state: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		message := &messages.Message{
			ClientID: 0,
			Type:     messages.MessageTypeServerGameUpdate,
			Payload:  payload,
		}

		if client.UDPAddress == nil {
			log.Trace("Client %d does not have a UDP address", client.ID)
			continue
		}

		err := network.WriteMessageToUDP(w.clientManager.GetUDPConn(), client.UDPAddress, message)
		if err != nil {
			log.Error("Failed to write message to UDP connection for client %d: %v", client.ID, err)
			continue
		}
	}

	return nil
}

func (w *BroadcastMessageWorker) handleServerPlayerUpdate(msg BroadcastMessage) error {
	playerUpdate, ok := msg.Message.(*messages.ServerPlayerUpdate)
	if !ok {
		return fmt.Errorf("failed to cast server player update message")
	}

	payload, err := messages.SerializeServerPlayerUpdate(playerUpdate)
	if err != nil {
		return fmt.Errorf("failed to serialize player state: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		message := &messages.Message{
			ClientID: 0,
			Type:     messages.MessageTypeServerPlayerUpdate,
			Payload:  payload,
		}

		if client.UDPAddress == nil {
			log.Trace("Client %d does not have a UDP address", client.ID)
			continue
		}

		err := network.WriteMessageToUDP(w.clientManager.GetUDPConn(), client.UDPAddress, message)
		if err != nil {
			log.Error("Failed to write message to UDP connection for client %d: %v", client.ID, err)
			continue
		}
	}

	return nil
}

func (w *BroadcastMessageWorker) handleServerNPCUpdate(msg BroadcastMessage) error {
	npcUpdate, ok := msg.Message.(*messages.ServerNPCUpdate)
	if !ok {
		return fmt.Errorf("failed to cast server NPC update message")
	}

	payload, err := messages.SerializeServerNPCUpdate(npcUpdate)
	if err != nil {
		return fmt.Errorf("failed to serialize NPC state: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		message := &messages.Message{
			ClientID: 0,
			Type:     messages.MessageTypeServerNPCUpdate,
			Payload:  payload,
		}

		if client.UDPAddress == nil {
			log.Trace("Client %d does not have a UDP address", client.ID)
			continue
		}

		err := network.WriteMessageToUDP(w.clientManager.GetUDPConn(), client.UDPAddress, message)
		if err != nil {
			log.Error("Failed to write message to UDP connection for client %d: %v", client.ID, err)
			continue
		}
	}

	return nil
}

func (w *BroadcastMessageWorker) handleServerNPCHit(msg BroadcastMessage) error {
	npcHit, ok := msg.Message.(*messages.ServerNPCHit)
	if !ok {
		return fmt.Errorf("failed to cast server NPC hit message")
	}

	payload, err := json.Marshal(npcHit)
	if err != nil {
		return fmt.Errorf("failed to marshal NPC hit message: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		msg := &messages.Message{
			ClientID: 0,
			Type:     messages.MessageTypeServerNPCHit,
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

func (w *BroadcastMessageWorker) handleServerNPCKill(msg BroadcastMessage) error {
	npcKill, ok := msg.Message.(*messages.ServerNPCKill)
	if !ok {
		return fmt.Errorf("failed to cast server NPC kill message")
	}

	payload, err := json.Marshal(npcKill)
	if err != nil {
		return fmt.Errorf("failed to marshal NPC kill message: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		msg := &messages.Message{
			ClientID: 0,
			Type:     messages.MessageTypeServerNPCKill,
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

func (w *BroadcastMessageWorker) handleServerPlayerHit(msg BroadcastMessage) error {
	playerHit, ok := msg.Message.(*messages.ServerPlayerHit)
	if !ok {
		return fmt.Errorf("failed to cast server player hit message")
	}

	payload, err := json.Marshal(playerHit)
	if err != nil {
		return fmt.Errorf("failed to marshal player hit message: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		msg := &messages.Message{
			ClientID: 0, // ClientID 0 means the message is from the server
			Type:     messages.MessageTypeServerPlayerHit,
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

func (w *BroadcastMessageWorker) handleServerPlayerKill(msg BroadcastMessage) error {
	playerKill, ok := msg.Message.(*messages.ServerPlayerKill)
	if !ok {
		return fmt.Errorf("failed to cast server player kill message")
	}

	payload, err := json.Marshal(playerKill)
	if err != nil {
		return fmt.Errorf("failed to marshal player kill message: %v", err)
	}

	for _, client := range w.clientManager.GetClients() {
		msg := &messages.Message{
			ClientID: 0,
			Type:     messages.MessageTypeServerPlayerKill,
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
