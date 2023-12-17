package servers

import (
	"encoding/json"
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

	s.ClientManager.SetUDPConn(udpConn)

	for {
		message, addr, err := ReadMessageFromUDP(udpConn)
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

// WriteMessageToUDP writes a Message to a UDP connection
func WriteMessageToUDP(conn *net.UDPConn, addr *net.UDPAddr, msg interface{}) error {
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
