package main

import (
	"flag"
	"log"
	"net/http"
)

var clientManager ClientManager = make(ClientManager)

var addr = flag.String("addr", "localhost:8081", "http service address")

type Recv_message struct {
	message string
	data    []byte
}

func echo(w http.ResponseWriter, r *http.Request) {
	nb_client := len(clientManager)
	client, err := NewClient(nb_client, w, r)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	if clientManager.has(client.name) {
		log.Fatal("Client already exist in clientManager")
	}

	clientManager.addClient(client)

	go client.HandleConnection(clientManager)

}

func main() {
	log.Println("Hello!!!!")
	flag.Parse()
	log.SetFlags(0)

	upgrader.CheckOrigin = checkOrigin

	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
