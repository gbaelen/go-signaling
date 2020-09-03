package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

var checkOriginTestCases = []struct {
	ok bool
	r  *http.Request
}{
	{false, &http.Request{Host: "example.org", Header: map[string][]string{"Origin": {"https://example.org"}}}},
	{false, &http.Request{Host: "github.io", Header: map[string][]string{"Origin": {"https://github.io"}}}},
	{false, &http.Request{Host: "autre.github.io", Header: map[string][]string{"Origin": {"https://autre.github.io"}}}},
	{false, &http.Request{Host: "localhost", Header: map[string][]string{"Origin": {"https://localhost"}}}},
	{false, &http.Request{Host: "localhost:8081", Header: map[string][]string{"Origin": {"https://localhost:8081"}}}},

	{true, &http.Request{Host: "localhost:8080", Header: map[string][]string{"Origin": {"http://localhost:8080"}}}},
	{true, &http.Request{Host: "localhost:8080", Header: map[string][]string{"Origin": {"https://localhost:8080"}}}},
	{true, &http.Request{Host: "gbaelen.github.io", Header: map[string][]string{"Origin": {"https://gbaelen.github.io/webrtc-video"}}}},
	{true, &http.Request{Host: "gbaelen.github.io", Header: map[string][]string{"Origin": {"https://gbaelen.github.io/"}}}},
	{true, &http.Request{Host: "gbaelen.github.io", Header: map[string][]string{"Origin": {"https://gbaelen.github.io"}}}},
}

var websocketMessagesOneUserTestCases = []struct {
	expected Message
	message  string
}{
	{Message{Type: Connected, Data: nil}, "{\"message\": \"connect\"}"},
	{Message{Type: NoPeer, Data: nil}, "{\"message\": \"offer\"}"},
	{Message{Type: UserList, Data: []string{"Client0"}}, "{\"message\": \"get_connected_users\"}"},
	{Message{Type: EmptyUser, Data: nil}, "{\"message\": \"call\"}"},
	{Message{Type: InvalidUser, Data: nil}, "{\"message\": \"call\", \"data\":\"Client0\"}"},
	{Message{Type: UserNotFound, Data: nil}, "{\"message\": \"call\", \"data\":\"client0\"}"},
	{Message{Type: NoPeer, Data: nil}, "{\"message\": \"answer\"}"},
	{Message{Type: NoPeer, Data: nil}, "{\"message\": \"candidate\"}"},
}

var websocketMessagesTwoUserNoPeerTestCases = []struct {
	expected Message
	message  string
}{
	{Message{Type: Connected, Data: nil}, "{\"message\": \"connect\"}"},
	{Message{Type: NoPeer, Data: nil}, "{\"message\": \"offer\"}"},
	{Message{Type: UserList, Data: []string{"Client0", "Client1"}}, "{\"message\": \"get_connected_users\"}"},
	{Message{Type: EmptyUser, Data: nil}, "{\"message\": \"call\"}"},
	{Message{Type: InvalidUser, Data: nil}, "{\"message\": \"call\", \"data\": \"Client0\"}"},
	{Message{Type: UserNotFound, Data: nil}, "{\"message\": \"call\", \"data\":\"client0\"}"},
	{Message{Type: NoPeer, Data: nil}, "{\"message\": \"answer\"}"},
	{Message{Type: NoPeer, Data: nil}, "{\"message\": \"candidate\"}"},
	{Message{Type: ReadyForConnection, Data: nil}, "{\"message\": \"call\", \"data\":\"Client1\"}"},
}

var websocketMessagesTwoUserAndPeerTestCases = []struct {
	expected Message
	message  string
}{}

func TestCheckOrigin(t *testing.T) {
	for _, test := range checkOriginTestCases {
		ok := checkOrigin(test.r)
		if test.ok != ok {
			t.Errorf("checkOrigin(%+v) returned %v, want %v", test.r, ok, test.ok)
		}
	}
}

func createWebsocketUser(t *testing.T, u string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	return ws
}

func createTestWebsocketServer(t *testing.T) (*httptest.Server, *websocket.Conn, string) {
	s := httptest.NewServer(http.HandlerFunc(echo))
	u := "ws" + strings.TrimPrefix(s.URL, "http")
	ws := createWebsocketUser(t, u)

	return s, ws, u
}

func cmpInterfaceAndStringSlice(i []interface{}, s []string) bool {
	if len(i) == len(s) {
		for i, v := range i {
			if s[i] != v {
				return false
			}
		}

		return true
	}

	return false
}

func executeTestWebsocketRequest(t *testing.T, ws *websocket.Conn, test_message string, expected Message) {
	if err := ws.WriteMessage(websocket.TextMessage, []byte(test_message)); err != nil {
		t.Fatalf("%v", err)
	}

	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var received_data Message
	err = json.Unmarshal(p, &received_data)
	if err != nil {
		log.Fatal("Couldn't parse received message to json: ", err)
	}

	if received_data.Type == UserList {
		if received_data.Type != expected.Type || !cmpInterfaceAndStringSlice(received_data.Data.([]interface{}), expected.Data.([]string)) {
			t.Fatalf("bad message")
		}
	} else {
		if received_data.Type != expected.Type || received_data.Data != expected.Data {
			t.Fatalf("For expected message type: %v and data: %v, got ==> type: %v, data: %v", expected.Type, expected.Data, received_data.Type, received_data.Data)
		}
	}

}

func TestWebsocket(t *testing.T) {
	s, ws, u := createTestWebsocketServer(t)
	log.Println(u)
	defer func(s *httptest.Server, ws *websocket.Conn) {
		s.Close()
		ws.Close()
	}(s, ws)

	t.Run("Test websocket default (no peer, 1 users)", func(t *testing.T) {
		for _, test := range websocketMessagesOneUserTestCases {
			executeTestWebsocketRequest(t, ws, test.message, test.expected)
		}
	})

	ws2 := createWebsocketUser(t, u)

	t.Run("Test websocket with no peer, 2 users", func(t *testing.T) {
		executeTestWebsocketRequest(t, ws2, "{\"message\": \"connect\"}", Message{Type: Connected, Data: nil})
		for _, test := range websocketMessagesTwoUserNoPeerTestCases {
			executeTestWebsocketRequest(t, ws, test.message, test.expected)
		}
	})

	t.Run("Test websocket with peer, 2 users", func(t *testing.T) {
		log.Println("Will check a lot of message for the webrtc connection")
	})

	//	ws3 := createWebsocketUser(t, u)
	t.Run("Test websocket with peer, 3 users", func(t *testing.T) {
		log.Println("Will check mainly the error for the second")
	})
}
