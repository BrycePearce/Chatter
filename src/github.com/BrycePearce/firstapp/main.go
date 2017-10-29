package main // main is the entry point of the application

// go run x y
import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // upgrades the http request to a websocket

// Hub is GO's hashmap, in this case it stores a map of all connected clients, and broadcasts messages to them all
type Hub struct {
	clients      map[*Client]bool
	broadcast    chan []byte
	addClient    chan *Client
	removeClient chan *Client
}

// initalize a new hub
var hub = Hub{
	broadcast:    make(chan []byte),
	addClient:    make(chan *Client),
	removeClient: make(chan *Client),
	clients:      make(map[*Client]bool),
}

// todo: figure out why two case conn
// Runs forever as a goroutine
func (hub *Hub) start() {
	for { // this for loop continously listens for messages
		// select is similar to switch/case
		select { // formally, select waits/blocks a message until one of its cases can run
		case conn := <-hub.addClient: // case for adding a client. <- adds 'hub.addClient'as an item to conn
			hub.clients[conn] = true
		case conn := <-hub.removeClient: // case for removing a client. <- adds 'hub.removeClient' as an item to conn
			if _, ok := hub.clients[conn]; ok {
				delete(hub.clients, conn) // The delete built-in function deletes the element with the specified key (m[key]) from the map. func delete(m map[Type]Type1, key Type)
				close(conn.send)          // close indicated that no more values will be sent on it https://gobyexample.com/closing-channels
			}
		case message := <-hub.broadcast: // case for sending a message to all clients
			for conn := range hub.clients { // for declares the loop, range iterates through the clients
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(hub.clients, conn)
				}
			}
		}
	}
}

// Pass data through this channel from the client functions below, which broadcasts the sent messages
type Client struct {
	ws *websocket.Conn // define this here so we can reference it in places such as write()
	// Hub passes broadcast messages to this channel
	send chan []byte // todo: figure out what this is doing. Think it just creates a byte array that can reallocate memory over and over, thus speeding things up a bit - http://openmymind.net/Introduction-To-Go-Buffered-Channels/
}

// Hub broadcasts a new message and this fires
func (c *Client) write() {
	// make sure to close the connection incase the loop exits
	defer func() { // defers this function from running until the parent function (write()) is done running. So this will only run if the for already exists, which would cause a problem
		c.ws.Close()
	}()

	for { // listening for messages
		select {
		case message, ok := <-c.send: // pass in message, ok from c.send (from Client passed into write()), and create a case using both variables
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{}) // http://www.gorillatoolkit.org/pkg/websocket#Conn.WriteMessage
				return
			}
			c.ws.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// New message received, so pass it to the Hub
func (c *Client) read() {
	defer func() {
		hub.removeClient <- c // define client as removeClient as defined in our Hub
		c.ws.Close()          // close the websocket connection
	}()

	for {
		_, message, err := c.ws.ReadMessage() // read the client message from the webscoket http://www.gorillatoolkit.org/pkg/websocket#Conn.ReadMessage
		if err != nil {                       // if there is an error, remove the client
			hub.removeClient <- c
			c.ws.Close() // close the websocket connection
			break
		}
		// otherwise, broadcast the message to the hub
		hub.broadcast <- message
	}
}

// wsPage handler creates a new client after upgrading the connection and storing it in the Hub.
func wsPage(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		http.NotFound(res, req)
		return
	}

	client := &Client{
		ws:   conn,
		send: make(chan []byte),
	}

	hub.addClient <- client

	go client.write()
	go client.read()
}

func homePage(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "../../../../index.html")
}

func main() {
	go hub.start()
	http.HandleFunc("/v1/ws", wsPage)
	http.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", nil)
}
