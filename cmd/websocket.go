package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	stop chan struct{}
}

func (c *Client) Initialize(url string) error {
	c.stop = make(chan struct{})
	var err error
	c.Conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ReadMessages() {
	for {
		select {
		case <-c.stop:
			return
		default:
			_, msg, err := c.Conn.ReadMessage()
			if err != nil {
				fmt.Println("Error reading message:", err)
				return
			}
			fmt.Printf("%s\n", msg)
		}
	}
}

func (c *Client) WriteMessages() {
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, _ := reader.ReadString('\n')
		if err := c.Conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			fmt.Println("Error writing message:", err)
			return
		}
	}
}

func (c *Client) HandleInterruptsAndPings() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.Conn.WriteMessage(websocket.PingMessage, []byte{})
		case <-interrupt:
			fmt.Println("signal: interrupt")
			c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			close(c.stop)
			return
		}
	}
}
