# This repo constains the minimal simulation of single user notification and multiple users broadcast using gorilla websocket

## Components Structure and Summary  

This app uses [go-gin](https://gin-gonic.com) web framework for http server and [gorrila/websocket](https://github.com/gorilla/websocket) websocket module for broadcasting and notification as the major packages. A simple html template is used to faciclitate client side real time nofication simulation. The html template is server side, parsed and rendered by `*gin.Engine` LoadHTMLGlob.

## File Structure

├── client.go \
├── go.mod \
├── go.sum \
├── hub.go \
├── main.go \
├── readme.md \
└── template \
    └── index.html

*main.go* : This the entry point and the main goroutine of the app. Gin engine is intantiated and served on port 8084. All http routes are implemented here.

*go.mod* : This is the module file which has metadata of used packages.

*template/index.html* : HTML template for establishing socket connection on the client side using JS web socket api `new WebSocket()`.

*client.go* : This is the go file with logic of reading messages from client's socket and sending message to back to the clients.

*hub.go* : This is the go file with broadcasting logic. It helps control how a client (purportedly from http request upgraded to websocket connection) is connected to the hub (group of clients or other websocket connections), it controls how a client is removed from the hub when websocket connection from the client side is closed, it also determines clients that new message would be broadcasted to.

### Operation flow

The basic flow of operation from initial page load to websocket connection on the client side, to pinging and sendingof message, to reading of the message on the server side socket, to broadcasting and sending of message back to the clients socket.

*Client and Hub struct:*

```
  type User struct {
    ID string
  }

  type CLient struct {
    hub  *Hub // Map of all connected clients and clients' signatures
    conn *websocket.Conn
    send chan []byte
    User User
  }

func (c *CLient) readPump() {}

func (c *CLient) writePump() {}

  type Hub struct {
    clients    map[*CLient]bool
    broadcast  chan []byte  // Incoming message from a client to other clients
    register   chan *CLient //Channel conveying a new client
    unregister chan *CLient //Channel conveying leaving client
  }

  func newHub() *Hub {
    return &Hub{
      broadcast:  make(chan []byte),
      register:   make(chan *CLient),
      unregister: make(chan *CLient),
      clients:    make(map[*CLient]bool),
    }
  }

  func (h *Hub) run() {}
```

When hhtp request is sent the servers `/` http route, the index page is loaded. On the index page, websocket is established and pointed to the `/broadcast` http route on which the request is upgraded to websocket connection.  A client address instance is created and added to the Hub through register channel.

Client's `readPump` reads incoming message from the client's side. The client's send channel is watched by the write pump and the message in []bytes is sent back to the clients side. That is the basic I/O flow of messages between the server and the frontend's socket. The Hub is the group of active clients and it helps during broadcasting. If broadcasting is needed, the readPump signals the Hub's broadcast channel, the message is read from the Hub's `broadcast` channel, all other clients in the Hub are signaled by sending the message through the individual client's `send` channel. The clients are `pointers`, the `writePump` method of each client process the message on the `send` channel and send the message back to the frontend's socket connection for each client.
