package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	authproviders "github.com/cbodonnell/flywheel/pkg/auth/providers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"nhooyr.io/websocket"
)

type NetworkManager struct {
	AuthProvider  authproviders.AuthProvider
	ClientManager *ClientManager
	MessageQueue  queue.Queue
	TCPServer     *TCPServer
	UDPServer     *UDPServer
	WSServer      *WSServer
}

type NewNetworkManagerOptions struct {
	AuthProvider  authproviders.AuthProvider
	ClientManager *ClientManager
	MessageQueue  queue.Queue
	TCPPort       int
	UDPPort       int
	WSPort        int
	WSServerTLS   *TLSConfig
}

func NewNetworkManager(options NewNetworkManagerOptions) *NetworkManager {
	return &NetworkManager{
		AuthProvider:  options.AuthProvider,
		ClientManager: options.ClientManager,
		MessageQueue:  options.MessageQueue,
		TCPServer: NewTCPServer(NewTCPServerOptions{
			Port: options.TCPPort,
		}),
		UDPServer: NewUDPServer(NewUDPServerOptions{
			Port: options.UDPPort,
		}),
		WSServer: NewWSServer(NewWSServerOptions{
			Port: options.WSPort,
			TLS:  options.WSServerTLS,
		}),
	}
}

func (n *NetworkManager) Start(ctx context.Context) {
	go n.TCPServer.Start(ctx, n.handleControlDisconnect, n.handleControlMessage)
	go n.UDPServer.Start(ctx, n.handleGameMessage)
	go n.WSServer.Start(ctx, n.handleControlDisconnect, n.handleControlMessage)
}

type ControlDisconnectHandler func(tcpConn net.Conn, wsConn *websocket.Conn)

func (n *NetworkManager) handleControlDisconnect(conn net.Conn, wsConn *websocket.Conn) {
	clientID := n.ClientManager.GetClientIDByTCPConn(conn)
	if clientID != 0 {
		n.ClientManager.DisconnectClient(clientID)
		log.Info("Client %d disconnected", clientID)
		return
	}

	clientID = n.ClientManager.GetClientIDByWSConn(wsConn)
	if clientID != 0 {
		n.ClientManager.DisconnectClient(clientID)
		log.Info("Client %d disconnected", clientID)
		return
	}

	log.Warn("Unknown client disconnected")
}

type ControlMessageHandler func(ctx context.Context, tcpConn net.Conn, wsConn *websocket.Conn, message *messages.Message)

func (n *NetworkManager) handleControlMessage(ctx context.Context, tcpConn net.Conn, wsConn *websocket.Conn, message *messages.Message) {
	if message.ClientID == 0 && message.Type != messages.MessageTypeClientLogin {
		log.Warn("Received message from unknown client that is not a login message")
		return
	}

	switch message.Type {
	case messages.MessageTypeClientLogin:
		clientID, err := n.handleClientLogin(ctx, tcpConn, wsConn, message)
		if err != nil {
			log.Error("Failed to handle client login: %v", err)
			if err := n.sendServerLoginFailure(ctx, clientID, err.Error()); err != nil {
				log.Error("Failed to send server login failure: %v", err)
			}
			return
		}
		log.Info("Client %d connected", clientID)
		if err := n.sendServerLoginSuccess(ctx, clientID); err != nil {
			log.Error("Failed to send server login success: %v", err)
		}
	case messages.MessageTypeClientSyncTime:
		if err := n.handleClientSyncTime(ctx, message); err != nil {
			log.Error("Failed to handle client sync time: %v", err)
		}
	default:
		if err := n.MessageQueue.Enqueue(message); err != nil {
			log.Error("Failed to enqueue message: %v", err)
		}
	}
}

// handleClientLogin handles a client login message.
func (n *NetworkManager) handleClientLogin(ctx context.Context, tcpConn net.Conn, wsConn *websocket.Conn, message *messages.Message) (uint32, error) {
	clientLogin := &messages.ClientLogin{}
	if err := json.Unmarshal(message.Payload, clientLogin); err != nil {
		return 0, fmt.Errorf("failed to unmarshal client login: %v", err)
	}

	token, err := n.AuthProvider.VerifyToken(ctx, clientLogin.Token)
	if err != nil {
		return 0, fmt.Errorf("failed to verify token: %v", err)
	}

	clientID, err := n.ClientManager.ConnectClient(tcpConn, wsConn, token.UID, clientLogin.CharacterID)
	if err != nil {
		return 0, fmt.Errorf("failed to connect client: %v", err)
	}

	return clientID, nil
}

func (n *NetworkManager) sendServerLoginSuccess(ctx context.Context, clientID uint32) error {
	serverLoginSuccess := &messages.ServerLoginSuccess{
		ClientID: clientID,
	}

	payload, err := json.Marshal(serverLoginSuccess)
	if err != nil {
		return fmt.Errorf("failed to marshal server login success: %v", err)
	}

	msg := &messages.Message{
		ClientID: 0,
		Type:     messages.MessageTypeServerLoginSuccess,
		Payload:  payload,
	}

	if err := n.SendReliableMessageToClient(ctx, clientID, msg); err != nil {
		return fmt.Errorf("failed to send server login success: %v", err)
	}

	return nil
}

func (n *NetworkManager) sendServerLoginFailure(ctx context.Context, clientID uint32, reason string) error {
	serverLoginFailure := &messages.ServerLoginFailure{
		Reason: reason,
	}

	payload, err := json.Marshal(serverLoginFailure)
	if err != nil {
		return fmt.Errorf("failed to marshal server login failure: %v", err)
	}

	msg := &messages.Message{
		ClientID: 0,
		Type:     messages.MessageTypeServerLoginFailure,
		Payload:  payload,
	}

	if err := n.SendReliableMessageToClient(ctx, clientID, msg); err != nil {
		return fmt.Errorf("failed to send server login failure: %v", err)
	}

	return nil
}

func (n *NetworkManager) handleClientSyncTime(ctx context.Context, message *messages.Message) error {
	clientSyncTime := &messages.ClientSyncTime{}
	if err := json.Unmarshal(message.Payload, clientSyncTime); err != nil {
		return fmt.Errorf("failed to unmarshal client sync time: %v", err)
	}

	serverSyncTime := &messages.ServerSyncTime{
		Timestamp:       time.Now().UnixMilli(),
		ClientTimestamp: clientSyncTime.Timestamp,
	}

	payload, err := json.Marshal(serverSyncTime)
	if err != nil {
		return fmt.Errorf("failed to marshal server sync time: %v", err)
	}

	msg := &messages.Message{
		ClientID: 0,
		Type:     messages.MessageTypeServerSyncTime,
		Payload:  payload,
	}

	if err := n.SendReliableMessageToClient(ctx, message.ClientID, msg); err != nil {
		return fmt.Errorf("failed to send server sync time: %v", err)
	}

	return nil
}

func (n *NetworkManager) handleGameMessage(ctx context.Context, addr *net.UDPAddr, message *messages.Message) {
	if message.ClientID == 0 {
		log.Warn("Received UDP message from unknown client, ignoring")
		return
	}

	if !n.ClientManager.Exists(message.ClientID) {
		log.Warn("Received UDP message from %d, but client is not connected", message.ClientID)
		return
	}

	switch message.Type {
	case messages.MessageTypeClientPing:
		if err := n.handleClientPing(ctx, message.ClientID, addr); err != nil {
			log.Error("Failed to handle client ping: %v", err)
		}
	default:
		if err := n.MessageQueue.Enqueue(message); err != nil {
			log.Error("Failed to enqueue message: %v", err)
		}
	}
}

func (n *NetworkManager) handleClientPing(ctx context.Context, clientID uint32, addr *net.UDPAddr) error {
	m := &messages.Message{
		ClientID: 0,
		Type:     messages.MessageTypeServerPong,
		Payload:  nil,
	}

	n.ClientManager.SetUDPAddress(clientID, addr)

	if err := n.SendUnreliableMessageToClient(ctx, clientID, m); err != nil {
		return fmt.Errorf("failed to write pong message to client: %v", err)
	}

	return nil
}

func (n *NetworkManager) SendUnreliableMessageToAll(ctx context.Context, msg *messages.Message) {
	for _, client := range n.ClientManager.GetClients() {
		if err := n.sendUnreliableMessageToClient(ctx, client, msg); err != nil {
			log.Error("Failed to send unreliable message to client %d: %v", client.ID, err)
		}
	}
}

func (n *NetworkManager) sendUnreliableMessageToClient(ctx context.Context, client *Client, msg *messages.Message) error {
	switch client.ConnectionType {
	case ClientConnectionTypeTCPUDP:
		if client.UDPAddress == nil {
			return fmt.Errorf("client %d does not have a UDP address", client.ID)
		}

		if err := WriteMessageToUDP(n.UDPServer.GetUDPConn(), client.UDPAddress, msg); err != nil {
			return fmt.Errorf("failed to write message to UDP connection for client %d: %v", client.ID, err)
		}
	case ClientConnectionTypeWebSocket:
		if err := WriteMessageToWS(ctx, client.WSConn, msg); err != nil {
			return fmt.Errorf("failed to write message to WebSocket connection for client %d: %v", client.ID, err)
		}
	default:
		return fmt.Errorf("unknown connection type for client %d: %v", client.ID, client.ConnectionType)
	}

	return nil
}

func (n *NetworkManager) SendUnreliableMessageToClient(ctx context.Context, clientID uint32, msg *messages.Message) error {
	client, err := n.ClientManager.GetClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client %d: %v", clientID, err)
	}

	if err := n.sendUnreliableMessageToClient(ctx, client, msg); err != nil {
		return fmt.Errorf("failed to send unreliable message to client %d: %v", clientID, err)
	}

	return nil
}

func (n *NetworkManager) SendReliableMessageToAll(ctx context.Context, msg *messages.Message) {
	for _, client := range n.ClientManager.GetClients() {
		if err := n.sendReliableMessageToClient(ctx, client, msg); err != nil {
			log.Error("Failed to send reliable message to client %d: %v", client.ID, err)
		}
	}
}

func (n *NetworkManager) sendReliableMessageToClient(ctx context.Context, client *Client, msg *messages.Message) error {
	switch client.ConnectionType {
	case ClientConnectionTypeTCPUDP:
		if err := WriteMessageToTCP(client.TCPConn, msg); err != nil {
			return fmt.Errorf("failed to write message to TCP connection for client %d: %v", client.ID, err)
		}
	case ClientConnectionTypeWebSocket:
		if err := WriteMessageToWS(ctx, client.WSConn, msg); err != nil {
			return fmt.Errorf("failed to write message to WebSocket connection for client %d: %v", client.ID, err)
		}
	default:
		return fmt.Errorf("unknown connection type for client %d: %v", client.ID, client.ConnectionType)
	}

	return nil
}

func (n *NetworkManager) SendReliableMessageToClient(ctx context.Context, clientID uint32, msg *messages.Message) error {
	client, err := n.ClientManager.GetClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client %d: %v", clientID, err)
	}

	if err := n.sendReliableMessageToClient(ctx, client, msg); err != nil {
		return fmt.Errorf("failed to send reliable message to client %d: %v", clientID, err)
	}

	return nil
}
