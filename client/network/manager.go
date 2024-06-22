package network

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cbodonnell/flywheel/client/ui"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

var (
	DefaultServerHostname = "localhost"
	DefaultServerTCPPort  = 8888
	DefaultServerUDPPort  = 8889
	DefaultWSServerPort   = 8890
)

// NetworkManager represents a network manager.
type NetworkManager struct {
	serverSettings       ServerSettings
	serverMessageQueue   queue.Queue
	serverConnectionType ServerConnectionType

	tcpClient        *TCPClient
	tcpClientErrChan chan error
	udpClient        *UDPClient
	udpClientErrChan chan error
	wsClient         *WSClient
	wsClientErrChan  chan error
	cancelClientCtx  context.CancelFunc

	clientWaitGroup sync.WaitGroup
	clientID        uint32
	clientIDMutex   sync.Mutex
	clientIDChan    <-chan uint32
	loginErr        error
	loginErrChan    <-chan error

	isConnected     bool
	serverTime      float64
	ping            float64
	deltaPing       float64
	recentRTTs      []int64
	serverTimeMutex sync.Mutex
	serverTimeChan  <-chan *messages.ServerSyncTime
}

type ServerSettings struct {
	Hostname string
	TCPPort  int
	UDPPort  int
	WSPort   int
}

type ServerConnectionType int

const (
	ServerConnectionTypeTCPUDP ServerConnectionType = iota
	ServerConnectionTypeWS
)

// NewNetworkManager creates a new network manager.
func NewNetworkManager(serverSettings ServerSettings, messageQueue queue.Queue) (*NetworkManager, error) {
	clientIDChan := make(chan uint32)
	loginErrChan := make(chan error)
	serverTimeChan := make(chan *messages.ServerSyncTime)

	m := &NetworkManager{
		serverSettings:     serverSettings,
		serverMessageQueue: messageQueue,
		clientIDChan:       clientIDChan,
		loginErrChan:       loginErrChan,
		serverTimeChan:     serverTimeChan,
	}

	if serverSettings.TCPPort != 0 && serverSettings.UDPPort != 0 {
		tcpClient := NewTCPClient(fmt.Sprintf("%s:%d", serverSettings.Hostname, serverSettings.TCPPort), messageQueue, clientIDChan, loginErrChan, serverTimeChan)
		udpClient, err := NewUDPClient(fmt.Sprintf("%s:%d", serverSettings.Hostname, serverSettings.UDPPort), messageQueue)
		if err != nil {
			return nil, fmt.Errorf("failed to create UDP client: %v", err)
		}
		m.serverConnectionType = ServerConnectionTypeTCPUDP
		m.tcpClient = tcpClient
		m.tcpClientErrChan = make(chan error)
		m.udpClient = udpClient
		m.udpClientErrChan = make(chan error)
	} else if serverSettings.WSPort != 0 {
		m.serverConnectionType = ServerConnectionTypeWS
		// TODO: dynamic websocket server URL
		wsClient := NewWSClient(fmt.Sprintf("ws://127.0.0.1:%d/", serverSettings.WSPort), messageQueue, clientIDChan, loginErrChan, serverTimeChan)
		m.wsClient = wsClient
		m.wsClientErrChan = make(chan error)
	} else {
		return nil, fmt.Errorf("no valid server ports provided")
	}

	return m, nil
}

// Start starts the network manager.
func (m *NetworkManager) Start(token string, characterID int32) error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelClientCtx = cancel

	switch m.serverConnectionType {
	case ServerConnectionTypeTCPUDP:
		if err := m.tcpClient.Connect(); err != nil {
			return fmt.Errorf("failed to start TCP client: %v", err)
		}

		m.clientWaitGroup.Add(1)
		go func(ctx context.Context) {
			defer m.clientWaitGroup.Done()
			if err := m.tcpClient.HandleMessages(ctx); err != nil {
				// TODO: find a way to do this without channels (maybe a flag w/ mutex?)
				m.tcpClientErrChan <- err
			}
		}(ctx)
	case ServerConnectionTypeWS:
		if err := m.wsClient.Connect(); err != nil {
			return fmt.Errorf("failed to start WebSocket client: %v", err)
		}

		m.clientWaitGroup.Add(1)
		go func(ctx context.Context) {
			defer m.clientWaitGroup.Done()
			if err := m.wsClient.HandleMessages(ctx); err != nil {
				m.wsClientErrChan <- err
			}
		}(ctx)
	default:
		return fmt.Errorf("unknown server connection type %d", m.serverConnectionType)
	}

	if err := m.login(token, characterID); err != nil {
		if strings.Contains(err.Error(), "is already connected") {
			return &ui.ActionableError{
				Message: "You are already connected to the server",
			}
		}
		return fmt.Errorf("failed to login: %v", err)
	}

	if err := m.startSyncTime(ctx); err != nil {
		return fmt.Errorf("failed to start time sync: %v", err)
	}

	if m.serverConnectionType == ServerConnectionTypeTCPUDP {
		if err := m.udpClient.Connect(); err != nil {
			return fmt.Errorf("failed to start UDP client: %v", err)
		}

		m.clientWaitGroup.Add(1)
		go func(ctx context.Context) {
			defer m.clientWaitGroup.Done()
			if err := m.udpClient.HandleMessages(ctx); err != nil {
				// TODO: find a way to do this without channels (maybe a flag w/ mutex?)
				m.udpClientErrChan <- err
			}
		}(ctx)

		if err := m.pingUDP(); err != nil {
			return fmt.Errorf("failed to ping UDP: %v", err)
		}
	}

	m.isConnected = true
	return nil
}

func (m *NetworkManager) login(token string, characterID int32) error {
	login := &messages.ClientLogin{
		Token:       token,
		CharacterID: characterID,
	}
	b, err := json.Marshal(login)
	if err != nil {
		return fmt.Errorf("failed to serialize login message: %v", err)
	}
	msg := &messages.Message{
		ClientID: 0,
		Type:     messages.MessageTypeClientLogin,
		Payload:  b,
	}
	if err := m.SendReliableMessage(msg); err != nil {
		return fmt.Errorf("failed to send login message: %v", err)
	}

	select {
	case m.clientID = <-m.clientIDChan:
		log.Info("Connected to server with client ID %d", m.clientID)
	case m.loginErr = <-m.loginErrChan:
		return fmt.Errorf("failed to login: %v", m.loginErr)
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timed out waiting for login response")
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
		serverTime := float64(serverSyncTime.Timestamp + rtt/2)

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

func (m *NetworkManager) setServerTime(serverTime float64, ping float64) {
	m.serverTimeMutex.Lock()
	defer m.serverTimeMutex.Unlock()
	m.deltaPing = ping - m.ping
	log.Trace("Server time sync - Before: %0.2f, After: %0.2f, Change: %0.2f, Delta Ping: %0.2f", m.serverTime, serverTime, serverTime-m.serverTime, m.deltaPing)
	m.serverTime = serverTime
	m.ping = ping
}

// UpdateServerTime updates the server time with the given delta time.
// This is intended to be called by the game update loop to keep
// the server time in sync with the client's time.
func (m *NetworkManager) UpdateServerTime(deltaTime float64) {
	m.serverTimeMutex.Lock()
	defer m.serverTimeMutex.Unlock()
	if m.serverTime == 0 {
		return
	}
	deltaTimeMs := deltaTime * 1000
	m.serverTime += deltaTimeMs + m.deltaPing
	m.deltaPing = 0
}

// Stop stops the network manager and its clients and clears the server message queue.
func (m *NetworkManager) Stop() error {
	if m.cancelClientCtx == nil {
		log.Warn("Network manager already stopped")
		return nil
	}
	m.cancelClientCtx()

	switch m.serverConnectionType {
	case ServerConnectionTypeTCPUDP:
		m.tcpClient.Close()
		m.udpClient.Close()
	case ServerConnectionTypeWS:
		m.wsClient.Close()
	default:
		return fmt.Errorf("unknown server connection type %d", m.serverConnectionType)
	}

	log.Debug("Waiting for clients to stop")
	m.clientWaitGroup.Wait()
	if err := m.serverMessageQueue.ClearQueue(); err != nil {
		return fmt.Errorf("failed to clear server message queue: %v", err)
	}

	m.clientID = 0
	m.cancelClientCtx = nil

	m.serverTimeMutex.Lock()
	defer m.serverTimeMutex.Unlock()
	m.serverTime = 0
	m.ping = 0
	m.deltaPing = 0
	m.recentRTTs = nil

	m.isConnected = false

	log.Info("Network manager stopped")

	return nil
}

func (m *NetworkManager) ServerSettings() ServerSettings {
	return m.serverSettings
}

func (m *NetworkManager) IsConnected() bool {
	return m.isConnected
}

func (m *NetworkManager) ServerTime() (serverTime float64, ping float64) {
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
	switch m.serverConnectionType {
	case ServerConnectionTypeTCPUDP:
		return m.tcpClient.SendMessage(msg)
	case ServerConnectionTypeWS:
		return m.wsClient.SendMessage(msg)
	default:
		return fmt.Errorf("unknown server connection type %d", m.serverConnectionType)
	}
}

func (m *NetworkManager) SendUnreliableMessage(msg *messages.Message) error {
	switch m.serverConnectionType {
	case ServerConnectionTypeTCPUDP:
		return m.udpClient.SendMessage(msg)
	case ServerConnectionTypeWS:
		return m.wsClient.SendMessage(msg)
	default:
		return fmt.Errorf("unknown server connection type %d", m.serverConnectionType)
	}
}
