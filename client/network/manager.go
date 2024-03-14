package network

import (
	"fmt"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// TODO: Client-side network manager similar to https://github.com/cbodonnell/flywheel-client/blob/main/Assets/Scripts/NetworkManager.cs

const (
	DefaultServerHostname = "localhost"
	DefaultServerTCPPort  = 8888
	DefaultServerUDPPort  = 8889
)

// NetworkManager represents a network manager.
type NetworkManager struct {
	serverMessageQueue queue.Queue
	tcpClient          *TCPClient
	udpClient          *UDPClient
	isConnected        bool
	clientID           uint32
}

// NewNetworkManager creates a new network manager.
func NewNetworkManager(messageQueue queue.Queue) (*NetworkManager, error) {
	udpClient, err := NewUDPClient(fmt.Sprintf("%s:%d", DefaultServerHostname, DefaultServerUDPPort), messageQueue)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP client: %v", err)
	}

	return &NetworkManager{
		serverMessageQueue: messageQueue,
		tcpClient:          NewTCPClient(fmt.Sprintf("%s:%d", DefaultServerHostname, DefaultServerTCPPort), messageQueue),
		udpClient:          udpClient,
	}, nil
}

// Start starts the network manager.
func (m *NetworkManager) Start() error {
	clientIDChan := make(chan uint32)

	// Connect to the server via TCP.
	go func(clientIDChan chan uint32) {
		err := m.tcpClient.Connect(clientIDChan)
		if err != nil {
			log.Error("Failed to start TCP client: %v", err)
		}
	}(clientIDChan)

	// Connect to the server via UDP.
	go func() {
		err := m.udpClient.Connect()
		if err != nil {
			log.Error("Failed to start UDP client: %v", err)
		}
	}()

	// TODO: make this change if either of the clients fail to start
	m.isConnected = true

	m.clientID = <-clientIDChan
	log.Info("Connected to server with client ID %d", m.clientID)

	pingUDPMsg := &messages.Message{
		ClientID: m.clientID,
		Type:     messages.MessageTypeClientPing,
	}
	if err := m.SendUnreliableMessage(pingUDPMsg); err != nil {
		return fmt.Errorf("failed to send UDP ping message: %v", err)
	}

	return nil
}

func (m *NetworkManager) ServerMessageQueue() queue.Queue {
	return m.serverMessageQueue
}

func (m *NetworkManager) SendReliableMessage(msg *messages.Message) error {
	return m.tcpClient.SendMessage(msg)
}

func (m *NetworkManager) SendUnreliableMessage(msg *messages.Message) error {
	return m.udpClient.SendMessage(msg)
}
