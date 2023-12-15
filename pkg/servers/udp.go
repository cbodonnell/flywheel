package servers

import (
	"fmt"
	"net"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
)

// UDPServer represents a UDP server.
type UDPServer struct {
	ClientManager *clients.ClientManager
	MessageQueue  *queue.MemoryQueue
	Port          string
}

// NewUDPServer creates a new UDP server.
func NewUDPServer(clientManager *clients.ClientManager, messageQueue *queue.MemoryQueue, port string) *UDPServer {
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
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("UDP server listening on", udpAddr.String())

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer udpConn.Close()

	for {
		message, addr, err := messages.ReadMessageFromUDP(udpConn)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		if !s.ClientManager.Exists(message.ClientID) {
			fmt.Printf("Received UDP packet from %d, but client is not connected\n", message.ClientID)
			continue
		}

		s.ClientManager.SetUDPAddress(message.ClientID, addr)
		s.MessageQueue.Enqueue(message)
	}
}
