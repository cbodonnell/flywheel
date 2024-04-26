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
		log.Debug("TCP Connection closed for client %d", connectedClientID)
	}()

	for {
		message, err := ReadMessageFromTCP(conn)
		if err != nil {
			if _, ok := err.(*ErrConnectionClosed); ok {
				log.Debug("Client %d disconnected", connectedClientID)
				return
			}
			log.Error("Error reading TCP message from client %d: %v", connectedClientID, err)
			continue
		}

		if message.ClientID == 0 && message.Type != messages.MessageTypeClientLogin {
			log.Warn("Received message from unknown client that is not a login message")
			continue
		}

		// TODO: create "handler" functions for each message type
		switch message.Type {
		case messages.MessageTypeClientLogin:
			clientLogin := &messages.ClientLogin{}
			if err := json.Unmarshal(message.Payload, clientLogin); err != nil {
				log.Error("Failed to unmarshal client login: %v", err)
				continue
			}

			token, err := s.AuthProvider.VerifyToken(ctx, clientLogin.Token)
			if err != nil {
				log.Error("Failed to verify ID token: %v", err)
				continue
			}

			clientID, err := s.ClientManager.ConnectClient(conn, token.UID)
			if err != nil {
				log.Error("Failed to add client: %v", err)
				conn.Close()
				return
			}
			connectedClientID = clientID

			log.Info("Client %d connected as %s", clientID, token.UID)

			assignID := messages.ServerLoginSuccess{
				ClientID: clientID,
			}

			payload, err := json.Marshal(assignID)
			if err != nil {
				log.Error("Failed to marshal assignID: %v", err)
				return
			}

			// Send the client its ID
			message := &messages.Message{
				ClientID: 0,
				Type:     messages.MessageTypeServerLoginSuccess,
				Payload:  payload,
			}
			if err := WriteMessageToTCP(conn, message); err != nil {
				log.Error("Error writing TCP message of type %s to client %d: %v", message.Type, clientID, err)
				return
			}
		case messages.MessageTypeClientSyncTime:
			clientSyncTime := &messages.ClientSyncTime{}
			if err := json.Unmarshal(message.Payload, clientSyncTime); err != nil {
				log.Error("Failed to unmarshal client sync time: %v", err)
				continue
			}

			serverSyncTime := &messages.ServerSyncTime{
				Timestamp:       time.Now().UnixMilli(),
				ClientTimestamp: clientSyncTime.Timestamp,
			}

			payload, err := json.Marshal(serverSyncTime)
			if err != nil {
				log.Error("Failed to marshal server sync time: %v", err)
				continue
			}

			msg := &messages.Message{
				ClientID: 0,
				Type:     messages.MessageTypeServerSyncTime,
				Payload:  payload,
			}
			if err := WriteMessageToTCP(conn, msg); err != nil {
				log.Error("Failed to send server pong: %v", err)
			}
		default:
			if err := s.MessageQueue.Enqueue(message); err != nil {
				log.Error("Failed to enqueue message: %v", err)
			}
		}
	}
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
type ErrConnectionClosed struct{}

func (e *ErrConnectionClosed) Error() string {
	return "connection closed"
}

// ReadMessageFromTCP reads a Message from a TCP connection
func ReadMessageFromTCP(conn net.Conn) (*messages.Message, error) {
	buf := make([]byte, messages.TCPMessageBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return nil, &ErrConnectionClosed{}
		}
		return nil, fmt.Errorf("failed to read message from TCP connection: %v", err)
	}

	msg, err := messages.DeserializeMessage(buf[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, nil
}
