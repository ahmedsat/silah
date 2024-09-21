package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/ahmedsat/silah/global"
)

type Client struct {
	conn     net.Conn
	incoming chan global.Message
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:     conn,
		incoming: make(chan global.Message),
	}, nil
}

func (c *Client) SendMessage(message global.Message) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(append(messageJSON, '\n'))
	return err
}

func (c *Client) ReceiveMessages() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		var message global.Message
		if err := json.Unmarshal(scanner.Bytes(), &message); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			continue
		}
		c.incoming <- message
	}
	c.conn.Close()
	close(c.incoming)
}

func (c *Client) GetIncomingChannel() <-chan global.Message {
	return c.incoming
}

func (c *Client) Close() error {
	return c.conn.Close()
}
