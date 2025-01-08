package sockio

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	*websocket.Conn
	*Sockio
	ID     uuid.UUID
	events map[string]func(string)
}

func NewClient(conn *websocket.Conn, sockio *Sockio) *Client {
	return &Client{
		Conn:   conn,
		Sockio: sockio,
		ID:     uuid.New(),
		events: make(map[string]func(string)),
	}
}

func (client *Client) On(event string, callback func(message string)) {
	if _, ok := client.events[event]; !ok {
		client.events[event] = callback
	}
}

func (client *Client) Emit(event string, message string) {
	client.WriteJSON(Event{
		Name:    event,
		Message: message,
	})
}

func (client *Client) Broadcast(event string, message string) {
	for _, c := range client.Sockio.clients {
		if c.ID != client.ID {
			c.WriteJSON(Event{
				Name:    event,
				Message: message,
			})
		}
	}
}

func (client *Client) Trigger(event string, message string) {
	if callback, ok := client.events[event]; ok {
		callback(message)
	}
}
