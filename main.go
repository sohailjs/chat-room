package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]bool)
var broadCast = make(chan []byte)

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	clients[conn] = true
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(strconv.Itoa(len(clients)) + " clients connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Print(err)
			conn.Close()
			delete(clients, conn)
		}
		broadCast <- msg
	}

}

func receiveMsg() {
	for {
		msg := <-broadCast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Print(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/chat", handler)
	go receiveMsg()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
