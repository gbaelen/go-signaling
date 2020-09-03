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

type MessageType string

const (
	GetConnectedUsers  MessageType = "get_connected_users"
	UserList           MessageType = "users_list"
	UserNotFound       MessageType = "user_not_found"
	Call               MessageType = "call"
	Connect            MessageType = "connect"
	Connected          MessageType = "connected"
	ReadyForConnection MessageType = "ready_to_establish_connection"
	Offer              MessageType = "offer"
	Answer             MessageType = "answer"
	Candidate          MessageType = "candidate"
)

type Message struct {
	Type MessageType `json:"message_type"`
	Data interface{} `json:"data"`
}

func NewClient(client_index int, w http.ResponseWriter, r *http.Request) (*Client, error) {
	name := "Client" + strconv.Itoa(client_index)
	c, err := upgrader.Upgrade(w, r, nil)

	//Might add a queue later to communicate with the client manager instead of inserting it as parameter of HandleConnection
	return &Client{name: name, conn: c}, err
}

func (c *Client) HandleConnection(cm ClientManager) {
	defer func() {
		c.conn.Close()
		cm.removeClient(c.name)
	}()

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
		case GetConnectedUsers:
			keys := cm.getKeys()
			c.sendMessage(UserList, keys)
		case Call:
			//Not pretty will have to find a better way to handle json in golang or a better way of managing marshal depending on type maybe?
			user := received_data["data"].(string)
			if cm.has(user) {
				cm[user].peer = c
				c.peer = cm[user]
				c.sendMessage(ReadyForConnection, nil)
			} else {
				c.sendMessage(UserNotFound, nil)
			}
		case Connect:
			c.sendMessage(Connected, nil)
		case Offer:
			c.peer.sendMessage(Offer, received_data["data"])
		case Answer:
			c.peer.sendMessage(Answer, received_data["data"])
		case Candidate:
			c.peer.sendMessage(Candidate, received_data["data"])
		}
	}
}

func (c *Client) sendMessage(message_type MessageType, data interface{}) {
	message, err := json.Marshal(&Message{Type: message_type, Data: data})
	if err != nil {
		log.Fatal("Error parsing message to []bytes")
	}

	c.conn.WriteMessage(1, message)
}

type ClientManager map[string]*Client

//Maybe merge addClient and NewClient together making the client Manager responsible for the creation of the client instance?
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
