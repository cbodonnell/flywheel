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
)

// TCPServer represents a TCP server.
type TCPServer struct {
	AuthProvider  authproviders.AuthProvider
	ClientManager *ClientManager
	MessageQueue  queue.Queue
	Port          int
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(authProvider authproviders.AuthProvider, clientManager *ClientManager, messageQueue queue.Queue, port int) *TCPServer {
	return &TCPServer{
		AuthProvider:  authProvider,
		ClientManager: clientManager,
		MessageQueue:  messageQueue,
		Port:          port,
	}
}

// Start starts the TCP server.
func (s *TCPServer) Start() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Error("Failed to resolve TCP address: %v", err)
		return
	}

	log.Info("TCP server listening on %s", tcpAddr.String())

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Error("Failed to listen on TCP address: %v", err)
		return
	}
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			log.Error("Failed to accept TCP connection: %v", err)
			continue
		}

		go s.handleTCPConnection(conn)
	}
}

// handleTCPConnection handles a TCP connection.
func (s *TCPServer) handleTCPConnection(conn net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())

	var connectedClientID uint32 // set after login

	defer func() {
		cancel()
		if connectedClientID != 0 {
			s.ClientManager.DisconnectClient(connectedClientID)
		}
		conn.Close()
		log.Info("Client %d disconnected", connectedClientID)
	}()

	for {
		message, err := ReadMessageFromTCP(conn)
		if err != nil {
			if _, ok := err.(*ErrConnectionClosed); ok {
				log.Trace("Connection closed for client %d: %v", connectedClientID, err)
				return
			}
			log.Error("Error reading TCP message from client %d: %v", connectedClientID, err)
			continue
		}

		if message.ClientID == 0 && message.Type != messages.MessageTypeClientLogin {
			log.Warn("Received message from unknown client that is not a login message")
			continue
		}

		switch message.Type {
		case messages.MessageTypeClientLogin:
			clientID, err := s.handleClientLogin(ctx, conn, message)
			if err != nil {
				log.Error("Failed to handle client login: %v", err)
				continue
			}
			connectedClientID = clientID
		case messages.MessageTypeClientSyncTime:
			if err := s.handleClientSyncTime(conn, message); err != nil {
				log.Error("Failed to handle client sync time: %v", err)
			}
		default:
			if err := s.MessageQueue.Enqueue(message); err != nil {
				log.Error("Failed to enqueue message: %v", err)
			}
		}
	}
}

// handleClientLogin handles a client login message.
func (s *TCPServer) handleClientLogin(ctx context.Context, conn net.Conn, message *messages.Message) (uint32, error) {
	clientLogin := &messages.ClientLogin{}
	if err := json.Unmarshal(message.Payload, clientLogin); err != nil {
		return 0, fmt.Errorf("failed to unmarshal client login: %v", err)
	}

	token, err := s.AuthProvider.VerifyToken(ctx, clientLogin.Token)
	if err != nil {
		return 0, fmt.Errorf("failed to verify token: %v", err)
	}

	clientID, err := s.ClientManager.ConnectClient(conn, token.UID, clientLogin.CharacterID)
	if err != nil {
		return 0, fmt.Errorf("failed to connect client: %v", err)
	}

	log.Info("Client %d connected as %s", clientID, token.UID)

	assignID := messages.ServerLoginSuccess{
		ClientID: clientID,
	}

	payload, err := json.Marshal(assignID)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal server login success: %v", err)
	}

	// Send the client its ID
	msg := &messages.Message{
		ClientID: 0,
		Type:     messages.MessageTypeServerLoginSuccess,
		Payload:  payload,
	}
	if err := WriteMessageToTCP(conn, msg); err != nil {
		return 0, fmt.Errorf("failed to send server login success: %v", err)
	}

	return clientID, nil
}

func (s *TCPServer) handleClientSyncTime(conn net.Conn, message *messages.Message) error {
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
	if err := WriteMessageToTCP(conn, msg); err != nil {
		return fmt.Errorf("failed to send server sync time: %v", err)
	}

	return nil
}

// WriteMessageToTCP writes a Message to a TCP connection
func WriteMessageToTCP(conn net.Conn, msg *messages.Message) error {
	b, err := messages.SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = conn.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write message to TCP connection: %v", err)
	}

	return nil
}

// ErrConnectionClosed is returned when the TCP connection is closed
type ErrConnectionClosed struct {
	Err error
}

func (e *ErrConnectionClosed) Error() string {
	return e.Err.Error()
}

// ReadMessageFromTCP reads a Message from a TCP connection
func ReadMessageFromTCP(conn net.Conn) (*messages.Message, error) {
	buf := make([]byte, messages.TCPMessageBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, &ErrConnectionClosed{err}
	}

	msg, err := messages.DeserializeMessage(buf[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}
