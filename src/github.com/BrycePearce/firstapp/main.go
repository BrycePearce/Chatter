package main // main is the entry point of the application

// go run x y
import (
	"net/http"
	"time"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // upgrades the http request to a websocket

func main() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "../../../../index.html")
	})

	// echo
	http.HandleFunc("/v1/ws", func(res http.ResponseWriter, req *http.Request) {
		var conn, _ = upgrader.Upgrade(res, req, nil) // create the websocket connection with (nil allows various headers to be passed back and forth between websocket) http://www.gorillatoolkit.org/pkg/websocket#Upgrader.Upgrade
		go func(conn *websocket.Conn) { // go routine
			for { // this for continuously listens for messages
			mType, msg, _ := conn.ReadMessage() // reads a message coming in from the client, takes the message/type of the message (types ex: ping/pong/text/binary)

			conn.WriteMessage(mType, msg) // write the message back to the client
			}
		}(conn)
	})

	// read the request message on GO side
	http.HandleFunc("/v2/ws", func(res http.ResponseWriter, req *http.Request) {
		var conn, _ = upgrader.Upgrade(res, req, nil)
		go func(conn *websocket.Conn) {
			for {
				_, msg, _ := conn.ReadMessage() // note: underscores are just placeholders for methods you aren't going to be using (mType, msg, err -> _, msg, _)
				println(string(msg))
			}
		}(conn)
	})

	// every 5 seconds, sends a message to the client
	http.HandleFunc("/v3/ws", func(res http.ResponseWriter, req *http.Request) {
		var conn, _ = upgrader.Upgrade(res, req, nil)
		go func(conn *websocket.Conn) {
			ch := time.Tick(5 *time.Second)

			for range ch { // send an instance 
				conn.WriteJSON(myStruct {
					Username: "Node231",
					FirstName: "Node",
					LastName: "231",
				})
			}
		}(conn)
	})

	// v3 only writes to socket. This version will recieve a message FROM the socket, in this case, a closing connection which we read 
	http.HandleFunc("/v4/ws", func(res http.ResponseWriter, req *http.Request) {
		var conn, _ = upgrader.Upgrade(res, req, nil)
		go func(conn *websocket.Conn) {
			for {
				_, _, err := conn.ReadMessage() // read all incoming messages, if it recieves a close message is recieved from the client, it will populate error object
				if err != nil { // so if error is populated
					conn.Close() // we close the connection, otherwise the connection will stay open forever and consume server resources
				}
			}
		}(conn)
		
		go func(conn *websocket.Conn) {
			ch := time.Tick(2 *time.Second)

			for range ch {
				conn.WriteJSON(myStruct {
					Username: "Node231",
					FirstName: "Node",
					LastName: "231",
				})
			}
		}(conn)
	})
	

	// serve it up
	http.ListenAndServe(":3000", nil)
}

type myStruct struct {
	Username string `json:"username"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
}