package network

import (
	"fmt"
	"math/rand"
	"net"
	"sync"

	"nhooyr.io/websocket"
)

const (
	// ClientIDMaxRetries represents the maximum number of retries when generating a unique ID
	ClientIDMaxRetries = 1024
	// ConnectionEventChannelSize represents the size of the client event channel
	ConnectionEventChannelSize = 1024
)

// Client represents a connected client
type Client struct {
	ID             uint32
	ConnectionType ClientConnectionType
	TCPConn        net.Conn
	WSConn         *websocket.Conn
	UDPAddress     *net.UDPAddr
	UserID         string
}

type ClientConnectionType int

const (
	ClientConnectionTypeTCPUDP ClientConnectionType = iota
	ClientConnectionTypeWebSocket
)

func (c *Client) Copy() *Client {
	copy := &Client{
		ID:             c.ID,
		ConnectionType: c.ConnectionType,
		TCPConn:        c.TCPConn,
		WSConn:         c.WSConn,
		UserID:         c.UserID,
	}
	if c.UDPAddress != nil {
		copy.UDPAddress = &net.UDPAddr{
			IP:   c.UDPAddress.IP,
			Port: c.UDPAddress.Port,
			Zone: c.UDPAddress.Zone,
		}
	}
	return copy
}

// ConnectionEvent represents an event that happened to a client
type ConnectionEvent struct {
	ClientID uint32
	Type     ConnectionEventType
	Data     interface{}
}

// ConnectionEventType represents the type of a client event
type ConnectionEventType int

const (
	ConnectionEventTypeConnect ConnectionEventType = iota
	ConnectionEventTypeDisconnect
)

type ClientConnectData struct {
	UserID      string
	CharacterID int32
}

// ClientManager manages connected clients
type ClientManager struct {
	clients             map[uint32]*Client
	clientUIDs          map[string]uint32
	clientsLock         sync.RWMutex
	connectionEventChan chan<- ConnectionEvent
}

// NewClientManager creates a new ClientManager
func NewClientManager(connectionEventChan chan<- ConnectionEvent) *ClientManager {
	return &ClientManager{
		clients:             make(map[uint32]*Client),
		clientUIDs:          make(map[string]uint32),
		connectionEventChan: connectionEventChan,
	}
}

// GetClients returns a slice with a copy of all connected clients.
func (cm *ClientManager) GetClients() []*Client {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	clients := make([]*Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client.Copy())
	}
	return clients
}

// GetClient returns a client by its ID
func (cm *ClientManager) GetClient(clientID uint32) (*Client, error) {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	client, ok := cm.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("client %d not found", clientID)
	}
	return client.Copy(), nil
}

// ConnectClient adds a new client to the manager and returns its ID
func (cm *ClientManager) ConnectClient(tcpConn net.Conn, wsConn *websocket.Conn, userID string, characterID int32) (uint32, error) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	if _, ok := cm.clientUIDs[userID]; ok {
		return 0, fmt.Errorf("user %s is already connected", userID)
	}

	clientID, err := cm.generateUniqueID(ClientIDMaxRetries)
	if err != nil {
		return 0, fmt.Errorf("failed to generate a unique ID: %v", err)
	}

	var connectionType ClientConnectionType
	if tcpConn != nil {
		connectionType = ClientConnectionTypeTCPUDP
	} else if wsConn != nil {
		connectionType = ClientConnectionTypeWebSocket
	} else {
		return 0, fmt.Errorf("neither TCP nor WebSocket connection provided")
	}

	client := &Client{
		ID:             clientID,
		ConnectionType: connectionType,
		TCPConn:        tcpConn,
		WSConn:         wsConn,
		UserID:         userID,
	}
	cm.clients[clientID] = client
	cm.clientUIDs[userID] = clientID

	event := ConnectionEvent{
		ClientID: clientID,
		Type:     ConnectionEventTypeConnect,
		Data: ClientConnectData{
			UserID:      userID,
			CharacterID: characterID,
		},
	}
	cm.connectionEventChan <- event

	return clientID, nil
}

// GetClientIDByTCPConn returns the ID of a client by its TCP connection.
// Returns 0 if the client is not found
func (cm *ClientManager) GetClientIDByTCPConn(conn net.Conn) uint32 {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	for _, client := range cm.clients {
		if client.TCPConn == conn {
			return client.ID
		}
	}
	return 0
}

// GetClientIDByWSConn returns the ID of a client by its WebSocket connection.
// Returns 0 if the client is not found
func (cm *ClientManager) GetClientIDByWSConn(conn *websocket.Conn) uint32 {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	for _, client := range cm.clients {
		if client.WSConn == conn {
			return client.ID
		}
	}
	return 0
}

// DisconnectClient removes a client from the manager
func (cm *ClientManager) DisconnectClient(clientID uint32) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	client, ok := cm.clients[clientID]
	if !ok {
		return
	}

	event := ConnectionEvent{
		ClientID: client.ID,
		Type:     ConnectionEventTypeDisconnect,
	}
	cm.connectionEventChan <- event

	delete(cm.clientUIDs, client.UserID)
	delete(cm.clients, clientID)
}

// SetUDPAddress sets the UDP address of a client
func (cm *ClientManager) SetUDPAddress(clientID uint32, addr *net.UDPAddr) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	// Don't update the UDP address if it's already set to the same value
	if cm.clients[clientID].UDPAddress != nil && cm.clients[clientID].UDPAddress.String() == addr.String() {
		return
	}

	cm.clients[clientID].UDPAddress = addr
}

func (cm *ClientManager) Exists(clientID uint32) bool {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	_, ok := cm.clients[clientID]
	return ok
}

// generateUniqueID generates a unique client ID with a maximum number of retries
// it reads from the clients, so it needs to be locked before calling
func (cm *ClientManager) generateUniqueID(maxRetries int) (uint32, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		id := rand.Uint32()
		if id == 0 {
			continue
		}
		if _, ok := cm.clients[id]; !ok {
			return id, nil
		}
	}

	return 0, fmt.Errorf("failed to generate a unique ID after %d attempts", maxRetries)
}
