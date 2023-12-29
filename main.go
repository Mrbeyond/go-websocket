package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	port           = ":8084"
	upgraderSingle = websocket.Upgrader{}
)

func main() {
	app := gin.Default()

	app.LoadHTMLGlob("template/*")

	hub := newHub()
	go hub.run()

	app.GET("/", HomeHandler)
	app.GET("/ws", BasicSocketHandler)
	app.GET("/board", func(ctx *gin.Context) {
		serverWs(hub, ctx.Writer, ctx.Request)
	})
	server := http.Server{
		Addr:           port,
		Handler:        app,
		ReadTimeout:    1 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	// go func() {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("listen:", err)
	} else {
		log.Printf(`up and runnning, serving on port%s`, port)
	}
	// }()
}

func BasicSocketHandler(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ws.Close()

	for {

		messageType, message, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("One", err.Error(), "here", messageType)
			log.Fatal(err)
			break
		}
		if messageType == -1 {
			fmt.Println("closing => ", messageType, string(message))
		} else {
			fmt.Println("messageType => ", messageType, string(message))
		}

		if string(message) == "ping" {
			message = []byte("pong")
		}
		err = ws.WriteMessage(messageType, message)
		if err != nil {
			fmt.Println("Two")
			log.Fatal(err)
			break
		}
	}
}

func HomeHandler(c *gin.Context) {
	type Data struct {
		Host string
	}

	data := Data{
		Host: c.Request.Host,
	}
	c.HTML(200, "index.html", data)
}
