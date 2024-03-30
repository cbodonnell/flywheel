package network

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// TCPServer represents a TCP server.
type TCPServer struct {
	ClientManager *ClientManager
	MessageQueue  queue.Queue
	Port          string
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(clientManager *ClientManager, messageQueue queue.Queue, port string) *TCPServer {
	return &TCPServer{
		ClientManager: clientManager,
		MessageQueue:  messageQueue,
		Port:          port,
	}
}

// Start starts the TCP server.
func (s *TCPServer) Start() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+s.Port)
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
	clientID, err := s.ClientManager.ConnectClient(conn)
	if err != nil {
		log.Error("Failed to add client: %v", err)
		conn.Close()
		return
	}

	defer func() {
		log.Debug("TCP Connection closed for client %d", clientID)
		s.ClientManager.DisconnectClient(clientID)
		conn.Close()
	}()

	log.Debug("TCP Connection established for client %d", clientID)

	assignID := messages.AssignID{
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
		Type:     messages.MessageTypeServerAssignID,
		Payload:  payload,
	}
	if err := WriteMessageToTCP(conn, message); err != nil {
		log.Error("Error writing TCP message of type %s to client %d: %v", message.Type, clientID, err)
		return
	}

	for {
		message, err := ReadMessageFromTCP(conn)
		if err != nil {
			if _, ok := err.(*ErrConnectionClosed); ok {
				log.Debug("Client %d disconnected", clientID)
				return
			}
			log.Error("Error reading TCP message from client %d: %v", clientID, err)
			continue
		}
		log.Trace("Received TCP message of type %s from client %d", message.Type, message.ClientID)

		switch message.Type {
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
	b, err := msg.Serialize()
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
	buf := make([]byte, messages.MessageBufferSize)
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
