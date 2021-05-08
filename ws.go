package main

import (
	"encoding/json"
	"errors"
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
	subjBet = "BET"

	subjPredictionStarted = "PREDICTION_STARTED"
	subjBetAccepted = "BET_ACCEPTED"
	subjPredictionChanged = "PREDICTION_CHANGED"
	subjGasInfo = "GAS_INFO"
)

var upgrader = websocket.Upgrader{}

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")

type Client struct {
	RemoteAddr string
	Username   string
	UserID     UID
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
var activePredictions map[uint64]*Prediction

func NewClient(remoteAddr string, uid UID, conn *websocket.Conn, db DB) (*Client, error) {
	client := &Client{}
	client.RemoteAddr = remoteAddr
	client.Username = "-" //TODO
	client.UserID = uid    //TODO
	client.Conn = conn
	client.db = db
	return client, nil
}

func (c *Client) HandleBet(message *Message) error {
	pIdStr, ok := message.Args["id"]
	if !ok {
		return errors.New("prediction Id is not set")
	}
	pId, err := strconv.ParseInt(pIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse prediction id: %s", err)
	}
	p, ok := activePredictions[uint64(pId)]
	if !ok {
		return fmt.Errorf("prediction with id %d not found", pId)
	}
	amount, err := strconv.ParseInt(message.Args["amount"], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse amount: %s", err)
	}
	bet := Bet{
		UserId:        UID(c.UserID),
		Amount:        uint64(amount),
		OnFirstOption: false,
	}
	err = p.AddBet(bet)
	if err != nil {
		return fmt.Errorf("can't add bet: %w", err)
	}
	c.Logf("bet accepted: %s", bet)
	msg := &Message{
		Subject: subjPredictionChanged,
		Args:    map[string]string{
			"Id": fmt.Sprintf("%d", p.Id),
		},
		Flags:   nil,
	}
	clientList.Broadcast(msg)
	msg.Subject = subjBetAccepted
	msg.Args = map[string]string{
		"predictionId": fmt.Sprintf("%d", p.Id),
	}
	return c.Conn.WriteJSON(msg)
}

func (c *Client) HandleStartPrediction(message *Message) error {
	//TODO: check permissions
	//TODO: validate input (opt1 != opt2, duplicate name, etc)
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
	activePredictions[p.Id] = p
	c.Logf("prediction started '%s' id:%d", p.Name, p.Id)
	msg := &Message{
		Subject: subjPredictionStarted,
		Args:    map[string]string{
			"name": message.Args["name"],
			"id": fmt.Sprintf("%d", p.Id),
			"opt1": p.Opt1,
			"opt2": p.Opt2,
			"delay": fmt.Sprintf("%d", p.StartDelaySeconds),
		},
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
	case subjBet:
		return c.HandleBet(message)
	}
	return fmt.Errorf("unknown msg subject: %s", message.Subject)
}

func (c *Client) SendGasInfo() error {
	gas, err := c.db.GetGasInfo(c.UserID)
	if err != nil {
		return err
	}
	return c.Conn.WriteJSON(Message{
		Subject: subjGasInfo,
		Args:    map[string]string{"gas": fmt.Sprintf("%d", gas)},
		Flags:   nil,
	})
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
	log.Printf("sending message to %d clients", len(cl.Clients))
	for _, c := range cl.Clients {
		err := c.Conn.WriteJSON(msg)
		c.Logf("sending: %s <- %+v", c, msg)
		if err != nil {
			c.Logf("broadcast: failed to send: %s", err)
		}
	}
	log.Printf("done")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo/", echo)

	activePredictions = map[uint64]*Prediction{}

	fs := http.FileServer(http.Dir("html"))
	var err error
	db, err = NewMySQLDB("defiler@/defiler?parseTime=true&loc=Local")
	if err != nil {
		log.Fatalf("unable to connect db: %s", err)
	}
	http.Handle("/", fs)
	log.Printf("Starting server at %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func getUID(r *http.Request) (UID, error) {
	uidRaw := r.URL.Query()["uid"]
	log.Printf("%+v", uidRaw)
	if len(uidRaw) == 0 {
		return -1, nil
	}
	uid, err := strconv.ParseInt(uidRaw[0], 10, 8)
	return UID(uid), err
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	uid, err := getUID(r)
	if err != nil {
		log.Println("can't get UID")
		return
	}
	client, err := NewClient(r.RemoteAddr, uid, c, db)
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
	err = client.SendGasInfo()
	if err != nil {
		return
	}
	var message Message
	for {
		message = Message{}
		mt, data, err := c.ReadMessage()
		if err != nil {
			e, ok := err.(*websocket.CloseError)
			if ok {
				client.Logf("disconnected: %s", e)
			} else {
				client.Logf("read error: %s")
			}
			break
		}
		if mt != websocket.TextMessage {
			client.Logf("non text message received: %d", mt)
			continue
		}
		err = json.Unmarshal(data, &message)
		if err != nil {
			client.Logf("unable to decode message: %s (%s)", err, data)
			continue
		}
		client.Logf("handling message %s", message)
		err = client.HandleMessage(&message)
		if err != nil {
			client.Logf("error handling message: %s", err)
			continue
		}
	}
}
