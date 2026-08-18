package websocket

import (
	"context"
	"log"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Client struct {
	conn *websocket.Conn

	send chan Message
	done chan struct{}

	closeOnce sync.Once
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan Message, 32),
		done: make(chan struct{}),
	}
}

func (c *Client) Start() {
	go c.writeLoop()
}

func (c *Client) writeLoop() {
	defer close(c.done)

	for message := range c.send {
		if err := wsjson.Write(
			context.Background(),
			c.conn,
			message,
		); err != nil {
			log.Printf(
				"WebSocket write failed: %v",
				err,
			)

			return
		}
	}
}

func (c *Client) Send(message Message) {
	select {
	case c.send <- message:

	default:
		log.Println("WebSocket client send buffer is full")
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
	})

	<-c.done

	_ = c.conn.Close(websocket.StatusNormalClosure, "")
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}
