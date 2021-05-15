package main

import (
	"fmt"
	"time"
)

type Message struct {
	Subject string            `json:"subject"`
	Args    map[string]string `json:"args"`
	Flags   []string          `json:"flags"`
}

func (m *Message) FillArgs(p *Prediction) {
	m.Args = map[string]string{
		"name":      p.Name,
		"id":        fmt.Sprintf("%d", p.Id),
		"opt1":      p.Opt1,
		"opt2":      p.Opt2,
		"delay":     fmt.Sprintf("%d", p.StartDelaySeconds),
		"createdAt": p.CreatedAt.Format(time.RFC3339),
		"startedAt": p.StartedAt.Format(time.RFC3339),
	}
}

func (m *Message) String() string {
	return fmt.Sprintf("%s %s %s", m.Subject, m.Args, m.Flags)
}