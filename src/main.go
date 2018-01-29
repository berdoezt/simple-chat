package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// class of Message and its properties
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// clients map, default value false, later will register with true value
var clients = make(map[*websocket.Conn]bool)
var broadcasts = make(chan Message)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func webSocketConn(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	// log error
	if err != nil {
		log.Fatal("Error on open websocket : ", err)
	}

	defer conn.Close()

	// register new client
	clients[conn] = true

	for {
		var m Message
		err := conn.ReadJSON(&m)
		if err != nil {
			log.Println("Error websocket: ", err)
			// if something error with the connection, remove the client from chat then exit
			delete(clients, conn)
			break
		}
		broadcasts <- m
	}
}

func handleMessages() {
	// endless loop
	for {
		//pass message from channel broadcasts
		msg := <-broadcasts
		//send msg object to every connected client in clients
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println("Error handlemessage: ", err)
				// if error when client exit, also remove it from clients
				client.Close()
				delete(clients, client)
				break
			}
		}
	}

}

func main() {
	// fileserver for every new client connect
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// handle incoming websocket connection
	http.HandleFunc("/ws", webSocketConn)

	// concurrent process with goroutine
	go handleMessages()

	// run the go server
	log.Println("Server run on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error on running server : ", err)
	}

}
