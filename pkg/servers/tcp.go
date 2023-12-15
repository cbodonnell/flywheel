// servers package

package servers

import (
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// TCPServer represents a TCP server.
type TCPServer struct {
	ClientManager *clients.ClientManager
	MessageQueue  *queue.MemoryQueue
	Port          string
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(clientManager *clients.ClientManager, messageQueue *queue.MemoryQueue, port string) *TCPServer {
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
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("TCP server listening on", tcpAddr.String())

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		go s.handleTCPConnection(conn)
	}
}

// handleTCPConnection handles a TCP connection.
func (s *TCPServer) handleTCPConnection(conn net.Conn) {
	fmt.Println("TCP Connection established")
	clientID, err := s.ClientManager.AddClient(conn)
	if err != nil {
		fmt.Printf("Error adding client: %v\n", err)
		conn.Close()
		return
	}

	defer func() {
		fmt.Printf("TCP Connection closed for client %d\n", clientID)
		s.ClientManager.RemoveClient(clientID)
		conn.Close()
	}()

	fmt.Printf("TCP Connection established for client %d\n", clientID)

	for {
		// TODO: handle a client disconnecting
		message, err := messages.ReadMessageFromTCP(conn)
		if err != nil {
			fmt.Printf("Error reading TCP message from client %d: %v\n", clientID, err)
			continue
		}
		fmt.Printf("Received TCP message of type %s from client %d: %x\n", message.Type, message.ClientID, message.Payload)
		s.MessageQueue.Enqueue(message)
	}
}
