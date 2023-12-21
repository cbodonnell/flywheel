package clients

import (
	"fmt"
	"net"
	"sync"
)

const (
	// ClientIDMaxRetries represents the maximum number of retries when generating a unique ID
	ClientIDMaxRetries = 1024
)

// Client represents a connected client
type Client struct {
	ID         uint32
	TCPConn    net.Conn
	UDPAddress *net.UDPAddr
}

// ClientManager manages connected clients
type ClientManager struct {
	clients     map[uint32]*Client
	clientsLock sync.RWMutex
	nextID      uint32
	// UDP connection for broadcasting to clients
	udpConn *net.UDPConn
}

// NewClientManager creates a new ClientManager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[uint32]*Client),
		nextID:  1,
	}
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

// GetClients returns a list of all connected clients
func (cm *ClientManager) GetClients() []*Client {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	clients := make([]*Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	return clients
}

// AddClient adds a new client to the manager and returns its ID
func (cm *ClientManager) AddClient(tcpConn net.Conn) (uint32, error) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()
	clientID, err := cm.GenerateUniqueID(ClientIDMaxRetries)
	if err != nil {
		return 0, fmt.Errorf("failed to generate a unique ID: %v", err)
	}
	client := &Client{
		ID:      clientID,
		TCPConn: tcpConn,
	}
	cm.clients[clientID] = client
	return clientID, nil
}

// RemoveClient removes a client from the manager.
func (cm *ClientManager) RemoveClient(clientID uint32) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	if _, exists := cm.clients[clientID]; exists {
		delete(cm.clients, clientID)
	}
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

// GetUDPAddress retrieves the UDP address of a client
func (cm *ClientManager) GetUDPAddress(clientID uint32) *net.UDPAddr {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	return cm.clients[clientID].UDPAddress
}

// GetClientByID retrieves a client by its ID
func (cm *ClientManager) GetClientByID(clientID uint32) *Client {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	return cm.clients[clientID]
}

func (cm *ClientManager) Exists(clientID uint32) bool {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	_, ok := cm.clients[clientID]
	return ok
}

// GenerateUniqueID generates a unique client ID with a maximum number of retries
// it reads from the clients, so it needs to be locked before calling
func (cm *ClientManager) GenerateUniqueID(maxRetries int) (uint32, error) {
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
