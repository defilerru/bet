package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
)

const (
	subjStartPrediction   = "CREATE_PREDICTION"
	subjPredictionStarted = "PREDICTION_STARTED"
)

var upgrader = websocket.Upgrader{}

var addr = flag.String("addr", "localhost:8080", "http service address")

type Client struct {
	RemoteAddr string
	Username   string
	UserID     int64
	Conn       *websocket.Conn

	db    DB
	index int
}

type ClientList struct {
	Clients []*Client
}

var clientList ClientList
var clientListMutex sync.Mutex
var db *MySQLDB

func NewClient(remoteAddr string, sid string, conn *websocket.Conn, db DB) (*Client, error) {
	client := &Client{}
	client.RemoteAddr = remoteAddr
	client.Username = "-" //TODO
	client.UserID = 42    //TODO
	client.Conn = conn
	client.db = db
	return client, nil
}

func (c *Client) HandleStartPrediction(message *Message) error {
	//TODO: check permissions
	//TODO: validate input (opt1 != opt2, etc)
	startDelaySeconds, err := strconv.ParseInt(message.Args["delay"], 10, 16)
	if err != nil {
		return err
	}
	p, err := CreatePrediction(message.Args["name"],
		message.Args["opt1"],
		message.Args["opt2"],
		uint16(startDelaySeconds),
		uint64(c.UserID),
		c.db)
	if err != nil {
		return err
	}
	c.Logf("prediction started '%s' id:%d", p.Name, p.Id)
	msg := &Message{
		Subject: subjPredictionStarted,
		Args:    map[string]string{"name": message.Args["name"], "id": fmt.Sprintf("%d", p.Id)},
		Flags:   nil,
	}
	go clientList.Broadcast(msg)
	return nil
}

func (c *Client) Logf(format string, args ...interface{}) {
	log.Printf("%s %s", c, fmt.Sprintf(format, args...))
}

func (c *Client) HandleMessage(message *Message) error {
	switch message.Subject {
	case subjStartPrediction:
		return c.HandleStartPrediction(message)
	}
	err := fmt.Errorf("unknown msg subject: %s", message.Subject)
	c.Logf("error handling message: %s", err)
	return err
}

func (c *Client) String() string {
	return fmt.Sprintf("%s %s[%d]", c.RemoteAddr, c.Username, c.UserID)
}

func (c *Client) Close() {
	err := c.Conn.Close()
	if err != nil {
		c.Logf("error closing conn: %s", err)
	}
	clientListMutex.Lock()
	defer clientListMutex.Unlock()
	length := len(clientList.Clients)
	clientList.Clients[length-1] = clientList.Clients[c.index]
	clientList.Clients = clientList.Clients[:length-1]
}

func (cl *ClientList) Push(client *Client) error {
	clientListMutex.Lock()
	defer clientListMutex.Unlock()
	cl.Clients = append(cl.Clients, client)
	client.index = len(cl.Clients) - 1
	return nil
}

func (cl *ClientList) Broadcast(msg *Message) {
	for _, c := range cl.Clients {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			c.Logf("broadcast: failed to send: %s", err)
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo/", echo)
	fs := http.FileServer(http.Dir("html"))
	var err error
	db, err = NewMySQLDB("defiler@/defiler?parseTime=true&loc=Local")
	if err != nil {
		log.Fatalf("unable to connect db: %s", err)
	}
	http.Handle("/", fs)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	client, err := NewClient(r.RemoteAddr, "todo", c, db)
	if err != nil {
		log.Printf("Failed to create client instance: %s (%s)", err, r.RemoteAddr)
		err = c.Close()
		if err != nil {
			log.Printf("Error closing ws connection: %s (%s)", err, r.RemoteAddr)
		}
		return
	}
	client.Logf("connected")
	defer client.Close()
	err = clientList.Push(client)
	if err != nil {
		return
	}
	message := &Message{}
	for {
		mt, data, err := c.ReadMessage()
		if err != nil {
			client.Logf("read error: ", err)
			break
		}
		if mt != websocket.TextMessage {
			client.Logf("non text message received: %d", mt)
			break
		}
		err = json.Unmarshal(data, message)
		if err != nil {
			client.Logf("unable to decode message: %s", err)
		}
		client.Logf("handling message %s", message)
		err = client.HandleMessage(message)
		if err != nil {
			client.Logf("error handling message:", err)
			break
		}
	}
}
