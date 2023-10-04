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
	stop chan bool
}

func (c *Client) Initialize(url string) error {
	c.stop = make(chan bool)
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
				if websocket.IsCloseError(err,
					websocket.CloseAbnormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNormalClosure) {
					fmt.Println("Connection Closed - Exiting")
				} else {
					fmt.Println("Error reading message:", err)
				}
				return
			}
			fmt.Printf("%s\n", msg)
		}
	}
}

func (c *Client) WriteMessages() {
	reader := bufio.NewReader(os.Stdin)
	inputChan := make(chan string)

	go func() {
		for {
			msg, _ := reader.ReadString('\n')
			inputChan <- msg
		}
	}()

	for {
		select {
		case <-c.stop:
			return
		case msg := <-inputChan:
			if err := c.Conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseAbnormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNormalClosure) {
					fmt.Println("Connection Closed - Exiting")
				} else {
					fmt.Println("Error writing message:", err)
				}
				return
			}
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
		case <-c.stop:
			return
		case <-ticker.C:
			c.Conn.WriteMessage(websocket.PingMessage, []byte{})
		case <-interrupt:
			fmt.Println("signal: interrupt")
			c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			c.stop <- true
			return
		}
	}
}
