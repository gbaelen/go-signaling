package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var clientManager ClientManager = make(ClientManager)

var port = os.Getenv("PORT")
var addr = flag.String("addr", "0.0.0.0:"+port, "http service address")

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

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I am running!")
}

func main() {
	log.Println("Hello!!!!")
	flag.Parse()
	log.SetFlags(0)

	upgrader.CheckOrigin = checkOrigin

	http.HandleFunc("/", home)
	http.HandleFunc("/echo", echo)
	log.Println("Server listening on: ", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
