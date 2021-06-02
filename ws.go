package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	subjStartPrediction = "CREATE_PREDICTION"
	subjBet             = "BET"

	subjPredictionStarted  = "PREDICTION_STARTED"
	subjBetAccepted        = "BET_ACCEPTED"
	subjPredictionChanged  = "PREDICTION_CHANGED"
	subjPredictionFinished = "PREDICTION_FINISHED"
	subjUserInfo           = "USER_INFO"
)

const canCreatePredictions = "CAN_CREATE_PREDICTIONS"

var upgrader = websocket.Upgrader{}

type Client struct {
	RemoteAddr string
	Username   string
	UserID     UID
	Balance    int64
	Conn       *websocket.Conn
	Moderator  bool

	db    DB
	index int
}

type ClientList struct {
	Clients []*Client
}

var clientList ClientList
var clientListMutex sync.Mutex
var db *MySQLDB
var activePredictions *Predictions

func NewClient(r *http.Request, db DB) (*Client, error) {
	var uid int64
	var err error
	client := &Client{}
	client.RemoteAddr = r.RemoteAddr
	client.db = db
	uidRaw := r.URL.Query()["uid"]
	if len(uidRaw) == 0 {
		uid = -1
	} else {
		uid, err = strconv.ParseInt(uidRaw[0], 10, 8)
	}
	if err != nil {
		uid = -1
	}
	client.UserID = UID(uid)
	err = db.GetUserInfo(UID(uid), &client.Balance, &client.Username, &client.Moderator)
	return client, err
}

func getPredictionByMsg(message *Message) (*Prediction, error) {
	pIdStr, ok := message.Args["id"]
	if !ok {
		return nil, errors.New("prediction Id is not set")
	}
	pId, err := strconv.ParseInt(pIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("can't parse prediction id: %s", err)
	}
	p, ok := activePredictions.PredictionIdMap[uint64(pId)]
	if !ok {
		return nil, fmt.Errorf("prediction with id %d not found", pId)
	}
	return p, nil
}

func (c *Client) HandlePredictionFinished(message *Message) error {
	var msg *Message
	p, err := getPredictionByMsg(message)
	if err != nil {
		return err
	}
	p.Opt1Won = message.Args["opt1Won"] == "true"
	err = c.db.EndPrediction(p)
	if err != nil {
		return err
	}
	activePredictions.Delete(p)
	msg = &Message{
		Subject: subjPredictionFinished,
		Args: map[string]string{
			"id": fmt.Sprintf("%d", p.Id),
			"opt1Won": fmt.Sprintf("%t", p.Opt1Won),
		},
		Flags:   []string{},
	}
	clientList.Broadcast(msg)
	return err
}

func (c *Client) HandleBet(message *Message) error {
	p, err := getPredictionByMsg(message)
	if err != nil {
		return err
	}
	amount, err := strconv.ParseInt(message.Args["amount"], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse amount: %s", err)
	}
	bet := Bet{
		UserId:        c.UserID,
		Amount:        uint64(amount),
		OnFirstOption: message.Args["opt1Win"] == "true",
	}
	err = p.AddBet(bet)
	if err != nil {
		return fmt.Errorf("can't add bet: %w", err)
	}
	c.Logf("bet accepted: %s", bet)
	msg := &Message{
		Subject: subjPredictionChanged,
		Args:    p.GetBetInfoArgs(),
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
	if !c.Moderator {
		return fmt.Errorf("attempted to start prediction: %+v", message)
	}
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
	activePredictions.Add(p)
	c.Logf("prediction started '%s' id:%d", p.Name, p.Id)
	msg := &Message{
		Subject: subjPredictionStarted,
		Flags:   nil,
	}
	msg.FillArgs(p)
	go p.WaitAndStopAccepting()
	go clientList.Broadcast(msg)
	return nil
}

func (c *Client) Logf(format string, args ...interface{}) {
	log.Printf("[%s] %s %s", time.Now().Format(time.UnixDate), c, fmt.Sprintf(format, args...))
}

func (c *Client) HandleMessage(message *Message) error {
	switch message.Subject {
	case subjStartPrediction:
		return c.HandleStartPrediction(message)
	case subjBet:
		return c.HandleBet(message)
	case subjPredictionFinished:
		return c.HandlePredictionFinished(message)
	}
	return fmt.Errorf("unknown msg subject: %s", message.Subject)
}

func (c *Client) SendActivePredictions() error {
	var err error
	var msg Message
	var msgUpdate Message
	msg.Subject = subjPredictionStarted
	msgUpdate.Subject = subjPredictionChanged
	for _, p := range activePredictions.Predictions {
		msg.FillArgs(p)
		err = c.Conn.WriteJSON(msg)
		if err != nil {
			c.Logf("error sending prediction: %s", err)
			return err
		}
		msgUpdate.Args = p.GetBetInfoArgs()
		err = c.Conn.WriteJSON(msgUpdate)
		if err != nil {
			c.Logf("error sending prediction update: %s", err)
			return err
		}
	}
	return nil
}

func (c *Client) SendUserInfo() error {
	flags := make([]string, 0)
	if c.Moderator {
		flags = append(flags, canCreatePredictions)
	}
	return c.Conn.WriteJSON(Message{
		Subject: subjUserInfo,
		Args: map[string]string{
			"gas": fmt.Sprintf("%d", c.Balance),
			"id":  fmt.Sprintf("%d", c.UserID),
		},
		Flags: flags,
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
	clientList.Clients[c.index] = clientList.Clients[length-1]
	clientList.Clients[c.index].index = c.index
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
	log.Printf("message: %+v", msg)
	for _, c := range cl.Clients {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			c.Logf("broadcast: failed to send: %s. Closing connection", err)
			c.Close()
		}
	}
	log.Printf("done")
}

func echo(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if !allowedOrigins.Contain(origin) {
		log.Printf("Suspicious request: %+v", r)
		log.Printf("Forbidden origin: %s. Disconnecting.", origin)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	client, err := NewClient(r, db)
	if err != nil {
		log.Printf("Failed to create client instance: %s (%s)", err, r.RemoteAddr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	client.Conn = c
	client.Logf("connected")
	defer client.Close()
	err = clientList.Push(client)
	if err != nil {
		return
	}
	err = client.SendUserInfo()
	if err != nil {
		return
	}
	err = client.SendActivePredictions()
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
