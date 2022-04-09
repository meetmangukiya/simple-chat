package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/cip8/autoname"
	"github.com/gorilla/websocket"
	"github.com/meetmangukiya/simple-chat/common"
)

type Config struct {
	host     string
	port     uint
	username string
	room     string
}

func parseFlags() Config {
	var host, username, room string
	var port uint

	flag.StringVar(&host, "host", "localhost", "server host")
	flag.StringVar(&username, "username", autoname.Generate(), "user to connecct as")
	flag.StringVar(&room, "room", "general", "room to connect to")
	flag.UintVar(&port, "port", 8080, "server port")
	flag.Parse()

	return Config{
		host: host, port: port, room: room, username: username,
	}
}

type Client struct {
	messagesIn  chan common.MessageParam
	messagesOut chan string
	conn        *websocket.Conn
}

func NewClient(u string) Client {
	log.Println("connecting", u)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("error occurred while connecting to room", err)
	}

	return Client{
		messagesIn:  make(chan common.MessageParam, 100),
		messagesOut: make(chan string, 100),
		conn:        c,
	}
}

func (c *Client) readTerminalMessages() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error occurred while reading input")
		}

		c.messagesOut <- line[:len(line)-1]
	}
}

func (c *Client) sendMessages() {
	for {
		msg := <-c.messagesOut
		params := common.MessageParam{
			Username: config.username,
			Text:     msg,
		}
		marshaledParams, err := json.Marshal(params)
		if err != nil {
			log.Fatal("error occurred while marshaling message")
		}

		message := common.Message{
			Op:     "sendMessage",
			Params: marshaledParams,
		}

		c.conn.WriteJSON(message)
	}
}

func (c *Client) receiveMessages() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Fatal("error occurred while reading message")
		}

		var message common.Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Fatal("error occurred while unmarshaleling websocket message")
		}

		params, err := message.ParseMessageParam()
		if err != nil {
			log.Fatal("error occurred while unmarshaleling message params")
		}

		c.messagesIn <- params
	}
}

func (c *Client) printIncomingMessages() {
	for {
		msg := <-c.messagesIn
		if msg.Username == config.username {
			continue
		}

		fmt.Printf("\r[%s] %s\n>> ", msg.Username, msg.Text)
	}
}

func (c *Client) Start() {
	var wg sync.WaitGroup
	wg.Add(4)
	go func() { c.printIncomingMessages(); wg.Done() }()
	go func() { c.readTerminalMessages(); wg.Done() }()
	go func() { c.sendMessages(); wg.Done() }()
	go func() { c.receiveMessages(); wg.Done() }()
	wg.Wait()
}

var config Config

func main() {
	config = parseFlags()

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", config.host, config.port), Path: fmt.Sprintf("/r/%s", config.room)}
	client := NewClient(u.String())
	client.Start()
}
