package main // main is the entry point of the application

import (
	"net/http"
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
			for {
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
				_, msg, _ := conn.ReadMessage()
				println(string(msg))
			}
		}(conn)
	})

	// serve it up
	http.ListenAndServe(":3000", nil)
}