package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// Client represents a connected client
type Client struct {
	ID        string
	Connected bool
	Conn      net.Conn
}

// ClientManager manages connected clients
type ClientManager struct {
	clients     map[string]*Client
	clientsLock sync.RWMutex
}

// NewClientManager creates a new ClientManager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*Client),
	}
}

// AddClient adds a new client to the manager
func (cm *ClientManager) AddClient(client *Client) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()
	cm.clients[client.ID] = client
}

// RemoveClient removes a client from the manager
func (cm *ClientManager) RemoveClient(clientID string) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()
	delete(cm.clients, clientID)
}

// GetClientByID retrieves a client by its ID
func (cm *ClientManager) GetClientByID(clientID string) *Client {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()
	return cm.clients[clientID]
}

// handleTCPConnection handles incoming TCP connections
func handleTCPConnection(conn net.Conn, clientManager *ClientManager) {
	clientID := generateUniqueID()
	client := &Client{
		ID:        clientID,
		Connected: true,
		Conn:      conn,
	}
	clientManager.AddClient(client)

	defer func() {
		fmt.Printf("TCP Connection closed for client %s\n", client.ID)
		client.Connected = false
		clientManager.RemoveClient(client.ID)
		conn.Close()
	}()

	fmt.Printf("TCP Connection established for client %s\n", client.ID)

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Client %s closed the connection.\n", client.ID)
				return
			}
			fmt.Printf("Error reading from client %s: %v\n", client.ID, err)
			continue
		}

		message = strings.TrimSpace(message)
		if message == "exit" {
			fmt.Printf("Client %s requested to close the connection.\n", client.ID)
			return
		}

		fmt.Printf("Received TCP message from client %s: %s\n", client.ID, message)
	}
}

// handleUDPPacket handles incoming UDP packets
func handleUDPPacket(buffer []byte, addr *net.UDPAddr, clientManager *ClientManager) {
	// Extract clientID from the UDP packet (replace with your logic)
	clientID := "123"

	// Use the clientID to check connection status
	if client := clientManager.GetClientByID(clientID); client != nil && client.Connected {
		message := string(buffer)
		fmt.Printf("Received UDP packet from %s (client %s): %s\n", addr.String(), clientID, message)
		// Implement your UDP communication logic here
	} else {
		fmt.Printf("Received UDP packet from %s, but client %s is not connected\n", addr.String(), clientID)
	}
}

// generateUniqueID generates a unique ID for a client (replace with your logic)
func generateUniqueID() string {
	return "123"
}

func startTCPServer(clientManager *ClientManager) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":8888")
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

		go handleTCPConnection(conn, clientManager)
	}
}

func startUDPServer(clientManager *ClientManager) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":8889")
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
		buffer := make([]byte, 1024)
		n, addr, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		go handleUDPPacket(buffer[:n], addr, clientManager)
	}
}

func main() {
	clientManager := NewClientManager()

	go startTCPServer(clientManager)
	go startUDPServer(clientManager)

	fmt.Println("Server started.")

	// Gracefully handle Ctrl+C to stop the program
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
	<-stopSignal

	// Perform cleanup or other graceful shutdown tasks here

	fmt.Println("Server shutting down gracefully.")
}
