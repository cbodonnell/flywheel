package clients

import (
	"fmt"
	"net"
	"sync"
)

const (
	// ClientIDMaxRetries represents the maximum number of retries when generating a unique ID
	ClientIDMaxRetries = 1024
	// ClientEventChannelSize represents the size of the client event channel
	ClientEventChannelSize = 1024
)

// Client represents a connected client
type Client struct {
	ID         uint32
	TCPConn    net.Conn
	UDPAddress *net.UDPAddr
}

// ClientEvent represents an event that happened to a client
type ClientEvent struct {
	ClientID uint32
	Type     ClientEventType
}

// ClientEventType represents the type of a client event
type ClientEventType int

const (
	ClientEventTypeConnect ClientEventType = iota
	ClientEventTypeDisconnect
)

// ClientManager manages connected clients
type ClientManager struct {
	clients     map[uint32]*Client
	clientsLock sync.RWMutex
	nextID      uint32
	// UDP connection for broadcasting to clients
	udpConn         *net.UDPConn
	clientEventChan chan ClientEvent
}

// NewClientManager creates a new ClientManager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:         make(map[uint32]*Client),
		nextID:          1,
		clientEventChan: make(chan ClientEvent, ClientEventChannelSize),
	}
}

// GetClientEventChan returns a one-way channel for receiving client events
func (cm *ClientManager) GetClientEventChan() <-chan ClientEvent {
	return cm.clientEventChan
}

// SetUDPConn sets the UDP listener connection for all clients
func (cm *ClientManager) SetUDPConn(conn *net.UDPConn) {
	cm.udpConn = conn
}

// GetUDPConn returns the UDP listener connection for all clients
func (cm *ClientManager) GetUDPConn() *net.UDPConn {
	if cm.udpConn == nil {
		panic("UDP connection is not set on ClientManager")
	}
	return cm.udpConn
}

// GetClients returns a slice with a copy of all connected clients.
func (cm *ClientManager) GetClients() []*Client {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	clients := make([]*Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		copy := &Client{
			ID:      client.ID,
			TCPConn: client.TCPConn,
		}
		if client.UDPAddress != nil {
			copy.UDPAddress = &net.UDPAddr{
				IP:   client.UDPAddress.IP,
				Port: client.UDPAddress.Port,
				Zone: client.UDPAddress.Zone,
			}
		}
		clients = append(clients, copy)
	}
	return clients
}

// ConnectClient adds a new client to the manager and returns its ID
func (cm *ClientManager) ConnectClient(tcpConn net.Conn) (uint32, error) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	clientID, err := cm.generateUniqueID(ClientIDMaxRetries)
	if err != nil {
		return 0, fmt.Errorf("failed to generate a unique ID: %v", err)
	}
	client := &Client{
		ID:      clientID,
		TCPConn: tcpConn,
	}
	cm.clients[clientID] = client

	event := ClientEvent{
		ClientID: clientID,
		Type:     ClientEventTypeConnect,
	}
	cm.clientEventChan <- event

	return clientID, nil
}

// DisconnectClient removes a client from the manager
func (cm *ClientManager) DisconnectClient(clientID uint32) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	event := ClientEvent{
		ClientID: clientID,
		Type:     ClientEventTypeDisconnect,
	}
	cm.clientEventChan <- event

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
		id := cm.nextID
		if _, ok := cm.clients[id]; !ok {
			cm.nextID++
			return id, nil
		}
		cm.nextID++
	}

	return 0, fmt.Errorf("failed to generate a unique ID after %d attempts", maxRetries)
}
