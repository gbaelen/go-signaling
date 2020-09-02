package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func checkOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}

	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	log.Println("Received websocket request from: ", u.Host)
	if u.Host == "localhost:8080" || u.Host == "gbaelen.github.io" {
		return true
	}

	return false
}

type Client struct {
	name string
	conn *websocket.Conn
	peer *Client
}

type Message struct {
	Type string      `json:"message_type"`
	Data interface{} `json:"data"`
}

func NewClient(client_index int, w http.ResponseWriter, r *http.Request) (*Client, error) {
	name := "Client" + strconv.Itoa(client_index)
	c, err := upgrader.Upgrade(w, r, nil)

	return &Client{name: name, conn: c}, err
}

func (c *Client) HandleConnection(cm ClientManager) {
	defer c.conn.Close()

	for {
		_, message, err := c.conn.ReadMessage()
		log.Println(string(message))
		if err != nil {
			log.Println("read:", err)
			break
		}

		var received_data map[string]interface{}
		err = json.Unmarshal(message, &received_data)
		if err != nil {
			log.Fatal("Couldn't parse received message to json: ", err)
		}

		switch received_data["message"] {
		case "get_connected":
			keys := cm.getKeys()
			c.sendMessage(&Message{Type: "connected_user", Data: keys})
		case "call":
			user := received_data["data"].(string)
			if cm.has(user) {
				log.Println(cm[user])
				cm[user].peer = c
				c.peer = cm[user]
				c.sendMessage(&Message{Type: "ready_to_establish_connection"})
			} else {
				c.sendMessage(&Message{Type: "user_not_found"})
			}
		case "connect":
			c.sendMessage(&Message{Type: "ok"})
		case "offer":
			log.Println(received_data["data"])
			c.peer.sendMessage(&Message{Type: "offer", Data: received_data["data"]})
		case "answer":
			c.peer.sendMessage(&Message{Type: "answer", Data: received_data["data"]})
		case "candidate":
			c.peer.sendMessage(&Message{Type: "candidate", Data: received_data["data"]})
		}
	}
}

func (c *Client) sendMessage(data *Message) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Error parsing message to []bytes")
	}
	c.conn.WriteMessage(1, message)
}

type ClientManager map[string]*Client

func (cm ClientManager) addClient(c *Client) {
	cm[c.name] = c
}

func (cm ClientManager) removeClient(name string) {
	delete(cm, name)
}

func (cm ClientManager) has(name string) bool {
	_, ok := cm[name]
	return ok
}

func (cm ClientManager) getKeys() []string {
	keys := make([]string, 0, len(cm))
	for k := range cm {
		keys = append(keys, k)
	}

	return keys
}
