package sockio

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Sockio struct {
	upgrader *websocket.Upgrader
	clients  map[uuid.UUID]*Client
	events   map[string]func(*Client)
}

type SockioConfig struct {
	HandshakeTimeout                time.Duration
	ReadBufferSize, WriteBufferSize int
	WriteBufferPool                 websocket.BufferPool
	Subprotocols                    []string
	Error                           func(w http.ResponseWriter, r *http.Request, status int, reason error)
	CheckOrigin                     func(r *http.Request) bool
	EnableCompression               bool
}

func New(config *SockioConfig) *Sockio {
	upgrader := &websocket.Upgrader{
		HandshakeTimeout:  config.HandshakeTimeout,
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		WriteBufferPool:   config.WriteBufferPool,
		Subprotocols:      config.Subprotocols,
		Error:             config.Error,
		CheckOrigin:       config.CheckOrigin,
		EnableCompression: config.EnableCompression,
	}

	return &Sockio{
		upgrader: upgrader,
		clients:  make(map[uuid.UUID]*Client),
		events:   make(map[string]func(*Client)),
	}
}

func (sockio *Sockio) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := sockio.upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	go sockio.handleConnection(conn)
}

func (sockio *Sockio) handleConnection(conn *websocket.Conn) {
	client := NewClient(conn, sockio)
	sockio.clients[client.ID] = client

	defer sockio.closeConnection(client)
	sockio.Trigger("connected", client)

	for {
		var event Event

		err := client.ReadJSON(&event)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Unexpected close error : ", err)
			}

			break
		}

		client.Trigger(event.Name, event.Message)
	}
}

func (sockio *Sockio) closeConnection(client *Client) {
	client.Close()
	delete(sockio.clients, client.ID)

	client.Trigger("disconnected", "")
}

func (sockio *Sockio) On(event string, callback func(*Client)) {
	if _, ok := sockio.events[event]; !ok {
		sockio.events[event] = callback
	}
}

func (sockio *Sockio) Trigger(event string, client *Client) {
	if callback, ok := sockio.events[event]; ok {
		callback(client)
	}
}
