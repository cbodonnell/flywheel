package network

import (
	"context"
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
)

// TCPServer represents a TCP server.
type TCPServer struct {
	port int
}

type NewTCPServerOptions struct {
	Port int
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(opts NewTCPServerOptions) *TCPServer {
	return &TCPServer{
		port: opts.Port,
	}
}

// Start starts the TCP server.
func (s *TCPServer) Start(ctx context.Context, messageChan chan<- *messages.Message, loginChan chan<- *LoginEvent, logoutChan chan<- *LogoutEvent) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", s.port))
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
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := tcpListener.Accept()
			if err != nil {
				log.Error("Failed to accept TCP connection: %v", err)
				continue
			}

			go s.handleTCPConnection(ctx, conn, messageChan, loginChan, logoutChan)
		}
	}
}

// handleTCPConnection handles a TCP connection.
func (s *TCPServer) handleTCPConnection(ctx context.Context, conn net.Conn, messageChan chan<- *messages.Message, loginChan chan<- *LoginEvent, logoutChan chan<- *LogoutEvent) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		logoutChan <- &LogoutEvent{
			TCPConn: conn,
		}
		conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			message, err := ReadMessageFromTCP(conn)
			if err != nil {
				if _, ok := err.(*ErrConnectionClosed); ok {
					log.Trace("Connection closed for %s", conn.RemoteAddr().String())
					return
				}
				log.Error("Error reading TCP message from %s: %v", conn.RemoteAddr().String(), err)
				continue
			}

			if message.ClientID != 0 {
				messageChan <- message
				continue
			}

			if message.Type == messages.MessageTypeClientLogin {
				loginChan <- &LoginEvent{
					TCPConn: conn,
					Message: message,
				}
				continue
			}

			log.Warn("Received a %s message from an unknown client", message.Type)
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
