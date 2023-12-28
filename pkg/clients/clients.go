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
	udpConn                *net.UDPConn
	connectEventManager    *ClientEventManager
	disconnectEventManager *ClientEventManager
}

// NewClientManager creates a new ClientManager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:                make(map[uint32]*Client),
		nextID:                 1,
		connectEventManager:    NewClientEventManager(),
		disconnectEventManager: NewClientEventManager(),
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

// GetClients returns a slice with a copy of all connected clients.
func (cm *ClientManager) GetClients() []*Client {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	clients := make([]*Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		copy := &Client{
			ID:      client.ID,
			TCPConn: client.TCPConn,
			UDPAddress: &net.UDPAddr{
				IP:   client.UDPAddress.IP,
				Port: client.UDPAddress.Port,
				Zone: client.UDPAddress.Zone,
			},
		}
		clients = append(clients, copy)
	}
	return clients
}

// AddClient adds a new client to the manager and returns its ID
func (cm *ClientManager) AddClient(tcpConn net.Conn) (uint32, error) {
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
	cm.connectEventManager.Trigger(ClientEvent{ClientID: clientID})
	return clientID, nil
}

// RemoveClient removes a client from the manager
func (cm *ClientManager) RemoveClient(clientID uint32) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()
	cm.disconnectEventManager.Trigger(ClientEvent{ClientID: clientID})
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

// RegisterConnectHandler registers a handler for client connect events.
// The handler will be called in a goroutine.
func (cm *ClientManager) RegisterConnectHandler(handler ClientEventHandler) {
	cm.connectEventManager.RegisterHandler(handler)
}

// RegisterDisconnectHandler registers a handler for client disconnect events.
// The handler will be called in a goroutine.
func (cm *ClientManager) RegisterDisconnectHandler(handler ClientEventHandler) {
	cm.disconnectEventManager.RegisterHandler(handler)
}
