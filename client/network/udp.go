package network

import (
	"context"
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
	log.Info("Connecting to UDP server at %s", c.serverAddr.String())
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP address: %v", err)
	}
	c.conn = conn
	return nil
}

func (c *UDPClient) HandleMessages(ctx context.Context) error {
	defer c.conn.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		b, err := ReadFromUDP(c.conn)
		if err != nil {
			if _, ok := err.(*ErrConnectionClosedByServer); ok {
				return err
			} else if _, ok := err.(*ErrConnectionClosedByClient); ok {
				log.Info("UDP connection closed by client")
				return nil
			}
			log.Error("Failed to read from UDP connection: %v", err)
			continue
		}
		go func() {
			if err := c.handleMessage(b); err != nil {
				log.Error("Failed to handle message: %v", err)
			}
		}()
	}
}

func (c *UDPClient) handleMessage(b []byte) error {
	msg, err := messages.DeserializeMessage(b)
	if err != nil {
		return fmt.Errorf("failed to deserialize message: %v", err)
	}

	switch msg.Type {
	case messages.MessageTypeServerPong:
		log.Debug("Received server pong")
	case messages.MessageTypeServerGameUpdate, messages.MessageTypeServerPlayerUpdate, messages.MessageTypeServerNPCUpdate:
		if err := c.messageQueue.Enqueue(msg); err != nil {
			return fmt.Errorf("failed to enqueue message: %v", err)
		}
	default:
		return fmt.Errorf("unknown message type: %v", msg.Type)
	}

	return nil
}

// Close closes the UDP connection.
func (c *UDPClient) Close() {
	if c.conn == nil {
		log.Warn("UDP connection is already closed")
		return
	}
	c.conn.Close()
}

// SendMessage sends a message to the UDP server.
func (c *UDPClient) SendMessage(msg *messages.Message) error {
	b, err := messages.SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = c.conn.WriteToUDP(b, c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to write message to UDP connection: %v", err)
	}

	return nil
}

// ReadFromUDP reads a buffer from a UDP connection
func ReadFromUDP(conn *net.UDPConn) ([]byte, error) {
	buf := make([]byte, messages.UDPMessageBufferSize)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		if err, ok := err.(*net.OpError); ok && err.Err.Error() == "use of closed network connection" {
			return nil, &ErrConnectionClosedByClient{}
		}
		return nil, fmt.Errorf("failed to read message from UDP connection: %v", err)
	}

	return buf[:n], nil
}
