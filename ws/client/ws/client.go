package ws

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client WebSocketの入出力とやりとりする
type Client struct {
	conn     *websocket.Conn
	done     bool
	Receiver chan []byte
	sender   chan []byte
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 20 * time.Second
	maxMessageSize = 16384
)

// New コンストラクタ
func New(conn *websocket.Conn) *Client {
	return &Client{
		conn:     conn,
		done:     false,
		Receiver: make(chan []byte, 128),
		sender:   make(chan []byte, 128),
	}
}

// Serve クライアントとして対応を開始する
func (c *Client) Serve(ctx context.Context, stopped chan struct{}) {
	defer close(stopped)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		c.startReading()
		wg.Done()
	}()
	go func() {
		c.startWriting(ctx)
		wg.Done()
	}()
	wg.Wait()

	c.done = true
}

// Destruct 送り口を閉じる
func (c *Client) Destruct() {
	close(c.sender)
}

// Send メッセージ送信のラッパー
func (c *Client) Send(message []byte) {
	if !c.done {
		c.sender <- message
	}
}

func (c *Client) startReading() {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	defer func() {
		c.conn.Close()
		close(c.Receiver)
	}()
	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			return
		}
		switch messageType {
		case websocket.TextMessage:
			c.Receiver <- message
		}
	}
}

func (c *Client) startWriting(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		case message, ok := <-c.sender:
			if !ok {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
