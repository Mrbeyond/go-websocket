package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline  = []byte{'\n'}
	space    = []byte{' '}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {

			userID := r.URL.Query().Get("user")
			log.Println(userID, "user_id joined")
			return true
		},
	}
)

type User struct {
	ID string
}

type CLient struct {
	hub  *Hub // Map of all connected clients and clients' signatures
	conn *websocket.Conn
	send chan []byte
	User User
}

func (c *CLient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Remove spaces around the message covering =>(replace newline at the beginning with space)
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// Data that is received and processed by hub.broadcast
		data := map[string][]byte{
			"message": message,
			"id":      []byte(c.User.ID),
		}
		userMessage, _ := json.Marshal(data)
		c.hub.broadcast <- userMessage
	}
}

func (c *CLient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				//Hub closes the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte("not ok"))
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Handles websocket request from peer
func serverWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &CLient{hub: hub, conn: conn, send: make(chan []byte, 265)}
	client.hub.register <- client
	client.User.ID = r.URL.Query().Get("user")
	go client.writePump()
	go client.readPump()
	client.send <- []byte("Hi " + client.User.ID + ", you are welcome")
}
