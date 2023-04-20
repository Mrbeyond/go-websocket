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
	hub.run()

	app.GET("/", HomeHandler)
	app.GET("/ws", BasicSocHandler)
	app.GET("/broad", func(ctx *gin.Context) {
		serverWs(hub, ctx.Writer, ctx.Request)
	})
	server := http.Server{
		Addr:           port,
		Handler:        app,
		ReadTimeout:    1 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 2 << 20,
	}

	// go func() {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("listen:", err)
	} else {
		log.Printf(`up and runnning, serving on port%s`, port)
	}
	// }()
}

func BasicSocHandler(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Zero")
		log.Fatal(err)
		return
	}

	defer ws.Close()

	for {

		messageType, message, err := ws.ReadMessage()
		if messageType == -1 {
			fmt.Println("closing =>", messageType, string(message))
		} else {
			fmt.Println("messageType =>", messageType, string(message))
		}
		if err != nil {
			fmt.Println("One", err.Error(), "here", messageType)
			log.Fatal(err)
			break
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

	fmt.Println(data, "ade")

	c.HTML(200, "index.html", data)
}
