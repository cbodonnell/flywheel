package network

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// UDPServer represents a UDP server.
type UDPServer struct {
	ClientManager *ClientManager
	MessageQueue  queue.Queue
	Port          string
}

// NewUDPServer creates a new UDP server.
func NewUDPServer(clientManager *ClientManager, messageQueue queue.Queue, port string) *UDPServer {
	return &UDPServer{
		ClientManager: clientManager,
		MessageQueue:  messageQueue,
		Port:          port,
	}
}

// Start starts the UDP server.
func (s *UDPServer) Start() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+s.Port)
	if err != nil {
		log.Error("Failed to resolve UDP address: %v", err)
		return
	}

	log.Info("UDP server listening on %s", udpAddr.String())

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Error("Failed to listen on UDP address: %v", err)
		return
	}
	defer udpConn.Close()

	s.ClientManager.SetUDPConn(udpConn)

	for {
		message, addr, err := ReadMessageFromUDP(udpConn)
		if err != nil {
			log.Error("Failed to read message from UDP connection: %v", err)
			continue
		}

		if !s.ClientManager.Exists(message.ClientID) {
			log.Warn("Received UDP message from %d, but client is not connected", message.ClientID)
			continue
		}

		log.Trace("Received UDP message of type %s from %d", message.Type, message.ClientID)

		// TODO: real identity verification
		s.ClientManager.SetUDPAddress(message.ClientID, addr)

		switch message.Type {
		case messages.MessageTypeClientPing:
			m := &messages.Message{
				ClientID: 0,
				Type:     messages.MessageTypeServerPong,
				Payload:  nil,
			}
			if err := WriteMessageToUDP(udpConn, addr, m); err != nil {
				log.Error("Failed to send server pong: %v", err)
			}
		default:
			if err := s.MessageQueue.Enqueue(message); err != nil {
				log.Error("Failed to enqueue message: %v", err)
			}
		}
	}
}

// WriteMessageToUDP writes a Message to a UDP connection
func WriteMessageToUDP(conn *net.UDPConn, addr *net.UDPAddr, msg *messages.Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = conn.WriteToUDP(jsonData, addr)
	if err != nil {
		return fmt.Errorf("failed to write message to UDP connection: %v", err)
	}

	return nil
}

// ReadMessageFromUDP reads a Message from a UDP connection
func ReadMessageFromUDP(conn *net.UDPConn) (*messages.Message, *net.UDPAddr, error) {
	jsonData := make([]byte, messages.MessageBufferSize)
	n, addr, err := conn.ReadFromUDP(jsonData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read message from UDP connection: %v", err)
	}

	msg := &messages.Message{}
	err = json.Unmarshal(jsonData[:n], msg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, addr, nil
}
