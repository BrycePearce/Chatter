package main // main is the entry point of the application

// go run x y
import (
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
		select {
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
