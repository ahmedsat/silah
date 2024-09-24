package server

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/ahmedsat/bayaan"
	"github.com/ahmedsat/silah/global"
)

type Server struct {
	clients              map[int]net.Conn
	clientID             int
	mu                   sync.Mutex
	clientMessageHandler ClientMessageHandler
}

type ClientMessageHandler interface {
	HandleMessage(clientID int, message global.Message) []global.Message
	ClientConnected(clientID int) []global.Message
	ClientDisconnected(clientID int) []global.Message
}

func NewServer(clientMessageHandler ClientMessageHandler) *Server {
	return &Server{
		clients:              make(map[int]net.Conn),
		clientID:             0,
		clientMessageHandler: clientMessageHandler,
	}
}

func (s *Server) Start(address string) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer ln.Close()

	bayaan.Info("Server started on %s", address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		s.mu.Lock()
		clientID := s.clientID
		s.clients[clientID] = conn
		s.clientID++
		s.mu.Unlock()

		go s.handleClient(clientID, conn)

		messages := s.clientMessageHandler.ClientConnected(clientID)
		s.broadcastMessages(messages)
	}
}

func (s *Server) handleClient(clientID int, conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	for {
		var message global.Message
		if err := decoder.Decode(&message); err != nil {
			break
		}

		messages := s.clientMessageHandler.HandleMessage(clientID, message)
		s.broadcastMessages(messages)
	}

	s.mu.Lock()
	delete(s.clients, clientID)
	s.mu.Unlock()

	messages := s.clientMessageHandler.ClientDisconnected(clientID)
	s.broadcastMessages(messages)
}

func (s *Server) broadcastMessages(messages []global.Message) {
	for _, message := range messages {
		messageJSON, err := json.Marshal(message)
		if err != nil {
			fmt.Printf("Error marshaling message: %v\n", err)
			continue
		}

		s.mu.Lock()
		for _, conn := range s.clients {
			_, err := conn.Write(append(messageJSON, '\n'))
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Server) SendMessage(clientID int, message ClientMessageHandler) error {
	s.mu.Lock()
	conn, ok := s.clients[clientID]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("client %d not found", clientID)
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = conn.Write(append(messageJSON, '\n'))
	return err
}
