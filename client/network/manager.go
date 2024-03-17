package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

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
	clientIDMutex      sync.Mutex
	clientIDChan       <-chan uint32
	serverTime         int64
	ping               float64
	recentRTTs         []int64
	serverTimeMutex    sync.Mutex
	serverTimeChan     <-chan *messages.ServerSyncTime
}

// NewNetworkManager creates a new network manager.
func NewNetworkManager(messageQueue queue.Queue) (*NetworkManager, error) {
	clientIDChan := make(chan uint32)
	serverTimeChan := make(chan *messages.ServerSyncTime)

	tcpClient := NewTCPClient(fmt.Sprintf("%s:%d", DefaultServerHostname, DefaultServerTCPPort), messageQueue, clientIDChan, serverTimeChan)
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
		clientIDChan:       clientIDChan,
		serverTimeChan:     serverTimeChan,
	}, nil
}

// Start starts the network manager.
func (m *NetworkManager) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelClientCtx = cancel

	// Connect to the server via TCP.
	m.clientWaitGroup.Add(1)
	go func(ctx context.Context) {
		defer m.clientWaitGroup.Done()
		err := m.tcpClient.Connect(ctx)
		if err != nil {
			m.tcpClientErrChan <- err
		}
	}(ctx)

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
	case m.clientID = <-m.clientIDChan:
		log.Info("Connected to server with client ID %d", m.clientID)
	}

	if err := m.startSyncTime(ctx); err != nil {
		return fmt.Errorf("failed to start time sync: %v", err)
	}

	if err := m.pingUDP(); err != nil {
		return fmt.Errorf("failed to ping UDP: %v", err)
	}

	return nil
}

func (m *NetworkManager) startSyncTime(ctx context.Context) error {
	if err := m.syncTime(); err != nil {
		return fmt.Errorf("failed to sync time: %v", err)
	}

	go func(ctx context.Context) {
		for {
			select {
			case <-time.After(5 * time.Second):
				if err := m.syncTime(); err != nil {
					log.Error("Failed to sync time: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return nil
}

func (m *NetworkManager) syncTime() error {
	clientSyncTime := &messages.ClientSyncTime{
		Timestamp: time.Now().UnixMilli(),
	}

	payload, err := json.Marshal(clientSyncTime)
	if err != nil {
		return fmt.Errorf("failed to marshal client sync time: %v", err)
	}

	msg := &messages.Message{
		ClientID: m.clientID,
		Type:     messages.MessageTypeClientSyncTime,
		Payload:  payload,
	}

	if err := m.SendReliableMessage(msg); err != nil {
		return fmt.Errorf("failed to send client sync time message: %v", err)
	}

	select {
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timed out waiting for server sync time message")
	case serverSyncTime := <-m.serverTimeChan:
		rtt := time.Now().UnixMilli() - serverSyncTime.ClientTimestamp
		serverTime := serverSyncTime.Timestamp + rtt/2
		log.Trace("Server time: %d, ping: %d", serverTime, rtt)

		// keep track of the last 10 RTTs to calculate an average ping
		m.recentRTTs = append(m.recentRTTs, rtt)
		for len(m.recentRTTs) > 10 {
			m.recentRTTs = m.recentRTTs[1:]
		}

		sampleRTTs := removeOutlierRTTs(m.recentRTTs)
		ping := 0.0
		for _, p := range sampleRTTs {
			ping += float64(p)
		}
		ping /= float64(len(sampleRTTs))

		m.setServerTime(serverTime, ping)
	}

	return nil
}

func (m *NetworkManager) pingUDP() error {
	pingUDPMsg := &messages.Message{
		ClientID: m.clientID,
		Type:     messages.MessageTypeClientPing,
	}
	return m.SendUnreliableMessage(pingUDPMsg)
}

func (m *NetworkManager) setServerTime(serverTime int64, ping float64) {
	m.serverTimeMutex.Lock()
	defer m.serverTimeMutex.Unlock()
	m.serverTime = serverTime
	m.ping = ping
}

// Stop stops the network manager and its clients and clears the server message queue.
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

func (m *NetworkManager) ServerTime() (serverTime int64, ping float64) {
	m.serverTimeMutex.Lock()
	defer m.serverTimeMutex.Unlock()
	return m.serverTime, m.ping
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
	m.clientIDMutex.Lock()
	defer m.clientIDMutex.Unlock()
	return m.clientID
}

func (m *NetworkManager) SendReliableMessage(msg *messages.Message) error {
	return m.tcpClient.SendMessage(msg)
}

func (m *NetworkManager) SendUnreliableMessage(msg *messages.Message) error {
	return m.udpClient.SendMessage(msg)
}
