package network

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// UDPClient represents a UDP client.
type UDPClient struct {
	serverAddr   *net.UDPAddr
	messageQueue queue.Queue
	conn         *net.UDPConn
}

// NewUDPClient creates a new UDP client.
func NewUDPClient(serverAddr string, messageQueue queue.Queue) (*UDPClient, error) {
	serverUDPAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	return &UDPClient{
		serverAddr:   serverUDPAddr,
		messageQueue: messageQueue,
	}, nil
}

// Connect starts the UDP client.
func (c *UDPClient) Connect() error {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP address: %v", err)
	}
	defer conn.Close()
	c.conn = conn

	for {
		msg, err := ReceiveUDPMessage(conn)
		if err != nil {
			fmt.Printf("Failed to receive message from UDP connection: %v\n", err)
			continue
		}
		log.Trace("Received message from UDP server of type %s", msg.Type)

		switch msg.Type {
		case messages.MessageTypeServerPong:
			log.Debug("Received server pong")
		case messages.MessageTypeServerGameUpdate:
			if err := c.messageQueue.Enqueue(msg); err != nil {
				log.Error("Failed to enqueue message: %v", err)
			}
		default:
			log.Warn("Received unexpected message type from UDP server: %s", msg.Type)
		}

	}
}

// SendMessage sends a message to the UDP server.
func (c *UDPClient) SendMessage(msg *messages.Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = c.conn.WriteToUDP(jsonData, c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to write message to UDP connection: %v", err)
	}

	return nil
}

// receiveMessages continuously receives messages from the UDP server.
func ReceiveUDPMessage(conn *net.UDPConn) (*messages.Message, error) {
	buffer := make([]byte, messages.MessageBufferSize)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from UDP connection: %v", err)
	}

	msg := &messages.Message{}
	err = json.Unmarshal(buffer[:n], msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return msg, nil
}
