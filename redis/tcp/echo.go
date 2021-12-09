package tcp

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"sync"
)

type Client struct {
	Conn    net.Conn
	Waiting sync.WaitGroup
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

type EchoHandler struct {
	activeConn sync.Map
	closing    bool
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing {
		_ = conn.Close()
	}

	client := &Client{Conn: conn}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("client disconnected")
				h.activeConn.Delete(client)
			} else {
				log.Println("error reading from client:", err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

func (h *EchoHandler) Close() error {
	log.Println("closing all active connections")
	h.closing = true
	h.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		_ = client.Close()
		return true
	})
	return nil
}
