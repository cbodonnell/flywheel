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
	tcpClient   *TCPClient
	udpClient   *UDPClient
	isConnected bool
}

// NewNetworkManager creates a new network manager.
func NewNetworkManager(messageQueue queue.Queue) *NetworkManager {
	return &NetworkManager{
		tcpClient: NewTCPClient(fmt.Sprintf("%s:%d", DefaultServerHostname, DefaultServerTCPPort), messageQueue),
		udpClient: NewUDPClient(fmt.Sprintf("%s:%d", DefaultServerHostname, DefaultServerUDPPort), messageQueue),
	}
}

// Start starts the network manager.
func (m *NetworkManager) Start() {
	// Connect to the server via TCP.
	go func() {
		err := m.tcpClient.Start()
		if err != nil {
			log.Error("Failed to start TCP client: %v", err)
		}
	}()

	// Connect to the server via UDP.
	go func() {
		err := m.udpClient.Start()
		if err != nil {
			log.Error("Failed to start UDP client: %v", err)
		}
	}()

	// TODO: make this change if either of the clients fail to start
	m.isConnected = true
}

func (m *NetworkManager) SendReliableMessage(msg *messages.Message) error {
	return m.tcpClient.SendMessage(msg)
}

func (m *NetworkManager) SendUnreliableMessage(msg *messages.Message) error {
	return m.udpClient.SendMessage(msg)
}
