package network

import (
	"context"
	"fmt"
	"sync"

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
	tcpClientErrChan   chan error
	udpClient          *UDPClient
	udpClientErrChan   chan error
	cancelClientCtx    context.CancelFunc
	clientWaitGroup    *sync.WaitGroup
	clientID           uint32
}

// NewNetworkManager creates a new network manager.
func NewNetworkManager(messageQueue queue.Queue) (*NetworkManager, error) {
	tcpClient := NewTCPClient(fmt.Sprintf("%s:%d", DefaultServerHostname, DefaultServerTCPPort), messageQueue)
	udpClient, err := NewUDPClient(fmt.Sprintf("%s:%d", DefaultServerHostname, DefaultServerUDPPort), messageQueue)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP client: %v", err)
	}

	return &NetworkManager{
		serverMessageQueue: messageQueue,
		tcpClient:          tcpClient,
		tcpClientErrChan:   make(chan error),
		udpClient:          udpClient,
		udpClientErrChan:   make(chan error),
		clientWaitGroup:    &sync.WaitGroup{},
	}, nil
}

// Start starts the network manager.
func (m *NetworkManager) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelClientCtx = cancel

	clientIDChan := make(chan uint32)

	// Connect to the server via TCP.
	m.clientWaitGroup.Add(1)
	go func(ctx context.Context, clientIDChan chan uint32) {
		defer m.clientWaitGroup.Done()
		err := m.tcpClient.Connect(ctx, clientIDChan)
		if err != nil {
			m.tcpClientErrChan <- err
		}
	}(ctx, clientIDChan)

	// Connect to the server via UDP.
	m.clientWaitGroup.Add(1)
	go func(ctx context.Context) {
		defer m.clientWaitGroup.Done()
		err := m.udpClient.Connect(ctx)
		if err != nil {
			m.udpClientErrChan <- err
		}
	}(ctx)

	select {
	case err := <-m.tcpClientErrChan:
		return fmt.Errorf("failed to start TCP client: %v", err)
	case m.clientID = <-clientIDChan:
		log.Info("Connected to server with client ID %d", m.clientID)
	}

	pingUDPMsg := &messages.Message{
		ClientID: m.clientID,
		Type:     messages.MessageTypeClientPing,
	}
	if err := m.SendUnreliableMessage(pingUDPMsg); err != nil {
		return fmt.Errorf("failed to send UDP ping message: %v", err)
	}

	return nil
}

// Stop stops the network manager.
func (m *NetworkManager) Stop() error {
	if m.cancelClientCtx == nil {
		log.Warn("Network manager already stopped")
		return nil
	}
	m.cancelClientCtx()

	m.tcpClient.Close()
	m.udpClient.Close()

	log.Debug("Waiting for clients to stop")
	m.clientWaitGroup.Wait()
	if err := m.serverMessageQueue.ClearQueue(); err != nil {
		return fmt.Errorf("failed to clear server message queue: %v", err)
	}

	m.clientID = 0
	m.cancelClientCtx = nil

	log.Info("Network manager stopped")

	return nil
}

func (m *NetworkManager) ServerMessageQueue() queue.Queue {
	return m.serverMessageQueue
}

func (m *NetworkManager) TCPClientErrChan() <-chan error {
	return m.tcpClientErrChan
}

func (m *NetworkManager) UDPClientErrChan() <-chan error {
	return m.udpClientErrChan
}

func (m *NetworkManager) ClientID() uint32 {
	return m.clientID
}

func (m *NetworkManager) SendReliableMessage(msg *messages.Message) error {
	return m.tcpClient.SendMessage(msg)
}

func (m *NetworkManager) SendUnreliableMessage(msg *messages.Message) error {
	return m.udpClient.SendMessage(msg)
}
