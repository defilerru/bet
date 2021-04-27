package main

import "encoding/json"

type Message struct {
	Subject string            `json:"subject"`
	Args    map[string]string `json:"args"`
	Flags   []string          `json:"flags"`
}

func Unserialize(data []byte) (*Message, error) {
	msg := &Message{}
	return msg, json.Unmarshal(data, msg)
}
