package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/meetmangukiya/simple-chat/common"
)

var upgrader = websocket.Upgrader{} // use default options

type ConnectionsMap struct {
	connections map[string][]*websocket.Conn
	mutexes     map[string]*sync.RWMutex
}

func (cm *ConnectionsMap) Mutex(room string) *sync.RWMutex {
	return cm.mutexes[room]
}

func (cm *ConnectionsMap) AddConnection(room string, c *websocket.Conn) {
	mutex := cm.Mutex(room)

	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := cm.connections[room]; !ok {
		cm.connections[room] = []*websocket.Conn{}
	}
	cm.connections[room] = append(cm.connections[room], c)

	log.Printf("added new connection to room: %s", room)
}

func (cm *ConnectionsMap) RemoveConnection(room string, c *websocket.Conn) {
	mutex := cm.Mutex(room)

	mutex.Lock()
	defer mutex.Unlock()

	newConnections := []*websocket.Conn{}

	for _, connection := range cm.connections[room] {
		if connection != c {
			newConnections = append(newConnections, connection)
		}
	}

	cm.connections[room] = newConnections
	log.Printf("removed a connection from room: %s", room)
}

func (cm *ConnectionsMap) InitRoom(room string) {
	isNew := false
	if _, ok := cm.connections[room]; !ok {
		cm.connections[room] = []*websocket.Conn{}
		isNew = true
	}
	if _, ok := cm.mutexes[room]; !ok {
		cm.mutexes[room] = &sync.RWMutex{}
		isNew = true
	}
	if isNew {
		log.Println("inited room", room)
	}
}

var connections = ConnectionsMap{connections: map[string][]*websocket.Conn{}, mutexes: map[string]*sync.RWMutex{}}

func broadcastMessage(room string, message string, username string) {
	mutex := connections.Mutex(room)
	mutex.RLock()
	defer mutex.RUnlock()

	params := common.MessageParam{
		Username: username,
		Text:     message,
	}
	marshaledParams, err := json.Marshal(params)
	if err != nil {
		fmt.Println("an error occurred while marshaling message params")
	}

	msg := common.Message{
		Op:     "sendMessage",
		Params: marshaledParams,
	}
	marshaled, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("an error occurred while marshaling message")
	}

	for _, conn := range connections.connections[room] {
		conn.WriteMessage(websocket.TextMessage, marshaled)
	}
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	room := vars["room"]
	connections.InitRoom(room)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// remove the connection from room-connections map on closing of connection
	defer func() { connections.RemoveConnection(room, c) }()

	log.Println("new user joined room", vars["room"])
	connections.AddConnection(room, c)

outer:
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		var msg common.Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("error occurred unmarshaling the message")
			break
		}

		switch msg.Op {
		case "sendMessage":
			messageParams, err := msg.ParseMessageParam()
			if err != nil {
				log.Println("error occurred while unmarshaling the param as message")
				break
			}
			log.Printf("user: %s sent message: %s\n", messageParams.Username, messageParams.Text)
			broadcastMessage(room, messageParams.Text, messageParams.Username)
		default:
			log.Println("unknown operation", msg.Op)
			break outer
		}

		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
