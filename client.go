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
	NoPeer             MessageType = "no_peer_selected"
	EmptyUser          MessageType = "wrong_parameter_empty_user_sent"
	InvalidUser        MessageType = "invalid_user_selected"
	Call               MessageType = "call"
	Connect            MessageType = "connect"
	Connected          MessageType = "connected"
	ReadyForConnection MessageType = "ready_to_establish_connection"
	Offer              MessageType = "offer"
	Answer             MessageType = "answer"
	Candidate          MessageType = "candidate"
)

type Message struct {
	Type MessageType `json:"message"`
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

		var received_data Message
		err = json.Unmarshal(message, &received_data)
		if err != nil {
			log.Fatal("Couldn't parse received message to json: ", err)
		}

		switch received_data.Type {
		case GetConnectedUsers:
			keys := cm.getKeys()
			c.sendMessage(UserList, keys)
		case Call:
			//Not pretty will have to find a better way to handle json in golang or a better way of managing marshal depending on type maybe?
			if received_data.Data != nil {
				user_name := received_data.Data.(string)
				if user_name != c.name {
					if cm.has(user_name) {
						cm[user_name].peer = c
						c.peer = cm[user_name]
						c.sendMessage(ReadyForConnection, nil)
					} else {
						c.sendMessage(UserNotFound, nil)
					}
				} else {
					c.sendMessage(InvalidUser, nil)
				}
			} else {
				c.sendMessage(EmptyUser, nil)
			}
		case Connect:
			c.sendMessage(Connected, nil)
		case Offer:
			if c.checkPeer() {
				c.peer.sendMessage(Offer, received_data.Data)
			}
		case Answer:
			if c.checkPeer() {
				c.peer.sendMessage(Answer, received_data.Data)
			}
		case Candidate:
			if c.checkPeer() {
				c.peer.sendMessage(Candidate, received_data.Data)
			}
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

func (c *Client) checkPeer() bool {
	if c.peer == nil {
		c.sendMessage(NoPeer, nil)
		return false
	}

	return true
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
