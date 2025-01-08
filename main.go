package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/rakhiazfa/simple-chat-app/sockio"
)

type User struct {
	*sockio.Client
	Name string
}

func NewUser(client *sockio.Client, name string) *User {
	return &User{
		Client: client,
		Name:   name,
	}
}

var users map[uuid.UUID]*User = make(map[uuid.UUID]*User)

type Message struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

func CreateMessage(from string, value string) string {
	message := Message{
		From:    from,
		Message: value,
	}

	jsonBytes, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}

	return string(jsonBytes)
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	io := sockio.New(&sockio.SockioConfig{})

	http.HandleFunc("/", fs.ServeHTTP)
	http.HandleFunc("/ws", io.Handler)

	io.On("connected", func(c *sockio.Client) {
		c.On("join", func(message string) {
			users[c.ID] = NewUser(c, message)

			c.Broadcast("message", CreateMessage("Server", fmt.Sprintf("%s has joined.", users[c.ID].Name)))
		})

		c.On("message", func(message string) {
			c.Broadcast("message", CreateMessage(users[c.ID].Name, message))
		})

		c.On("disconnected", func(message string) {
			c.Broadcast("message", CreateMessage("Server", fmt.Sprintf("%s leaves the room.", users[c.ID].Name)))
		})
	})

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", nil)
}
