package network

import (
	"context"
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
)

// UDPServer represents a UDP server.
type UDPServer struct {
	port int
	conn *net.UDPConn
}

// NewUDPServerOptions represents the options for creating a new UDP server.
type NewUDPServerOptions struct {
	Port int
}

// NewUDPServer creates a new UDP server.
func NewUDPServer(opts NewUDPServerOptions) *UDPServer {
	return &UDPServer{
		port: opts.Port,
	}
}

// Start starts the UDP server.
func (s *UDPServer) Start(ctx context.Context, messageChan chan<- *messages.Message, pingChan chan<- *PingEvent) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Error("Failed to resolve UDP address: %v", err)
		return
	}

	log.Info("UDP server listening on %s", udpAddr.String())

	s.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Error("Failed to listen on UDP address: %v", err)
		return
	}
	defer s.conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			message, addr, err := ReadMessageFromUDP(s.conn)
			if err != nil {
				log.Error("Failed to read message from UDP connection: %v", err)
				continue
			}

			if message.ClientID == 0 {
				log.Warn("Received UDP message from an unknown client")
				continue
			}

			if message.Type == messages.MessageTypeClientPing {
				pingChan <- &PingEvent{
					Addr:    addr,
					Message: message,
				}
				continue
			}

			messageChan <- message
		}
	}
}

func (s *UDPServer) GetUDPConn() *net.UDPConn {
	if s.conn == nil {
		panic("UDP server not started")
	}

	return s.conn
}

// WriteMessageToUDP writes a Message to a UDP connection
func WriteMessageToUDP(conn *net.UDPConn, addr *net.UDPAddr, msg *messages.Message) error {
	b, err := messages.SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	_, err = conn.WriteToUDP(b, addr)
	if err != nil {
		return fmt.Errorf("failed to write message to UDP connection: %v", err)
	}

	return nil
}

// ReadMessageFromUDP reads a Message from a UDP connection
func ReadMessageFromUDP(conn *net.UDPConn) (*messages.Message, *net.UDPAddr, error) {
	buf := make([]byte, messages.UDPMessageBufferSize)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read message from UDP connection: %v", err)
	}

	msg, err := messages.DeserializeMessage(buf[:n])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	return msg, addr, nil
}
