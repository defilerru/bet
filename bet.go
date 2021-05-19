package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type UID int64

type Duplicate struct{}

func (d *Duplicate) Error() string {
	return "duplicate"
}

type Insufficient struct{}

func (d *Insufficient) Error() string {
	return "insufficient"
}

type NotFound struct{}

func (d *NotFound) Error() string {
	return "not found"
}

type Prediction struct {
	Id   uint64
	Name string
	Opt1 string
	Opt2 string
	Bets map[UID]Bet

	Balance1 uint64
	Balance2 uint64

	CreatedBy uint64
	CreatedAt time.Time
	StartedAt time.Time
	FinishedAt time.Time
	Opt1Won bool

	StartDelaySeconds uint16

	started bool
	db      DB
}

type Bet struct {
	UserId        UID
	Amount        uint64
	OnFirstOption bool
}

func CreatePrediction(name string, opt1 string, opt2 string, startDelaySeconds uint16, createdBy uint64, db DB) (prediction *Prediction, err error) {
	prediction = &Prediction{
		CreatedBy: createdBy,
		Name: name,
		Opt1: opt1,
		Opt2: opt2,
		Bets: map[UID]Bet{},
		StartDelaySeconds: startDelaySeconds,
		db:   db,
	}
	return prediction, db.CreatePrediction(prediction)
}

func (p *Prediction) AddBet(bet Bet) error {
	if !p.StartedAt.IsZero() {
		return errors.New("too late")
	}
	_, ok := p.Bets[bet.UserId]
	if ok {
		return &Duplicate{}
	}
	err := p.db.CreateBet(p, &bet)
	if err != nil {
		return err
	}
	p.Bets[bet.UserId] = bet
	if bet.OnFirstOption {
		p.Balance1 += bet.Amount
	} else {
		p.Balance2 += bet.Amount
	}
	return nil
}

func (p *Prediction) CalculateInfo() map[string]string {
	var amountOpt1 uint64
	var amountOpt2 uint64
	var ppl1 uint32
	var ppl2 uint32
	var coef1 float32
	var coef2 float32
	var per1 float32
	var per2 float32
	for _, bet := range p.Bets {
		if bet.OnFirstOption {
			ppl1 += 1
			amountOpt1 += bet.Amount
		} else {
			ppl2 += 1
			amountOpt2 += bet.Amount
		}
	}
	total := float32(amountOpt1 + amountOpt2)
	coef1 = total / float32(amountOpt1)
	coef2 = total / float32(amountOpt2)
	per1 = float32(amountOpt1) / total
	per2 = float32(amountOpt2) / total
	return map[string]string{
		"amountOpt1": fmt.Sprintf("%d", amountOpt1),
		"amountOpt2": fmt.Sprintf("%d", amountOpt2),
		"ppl1": fmt.Sprintf("%d", ppl1),
		"ppl2": fmt.Sprintf("%d", ppl2),
		"coef1": fmt.Sprintf("1:%.2f", coef1),
		"coef2": fmt.Sprintf("1:%.2f", coef2),
		"per1": fmt.Sprintf("%.1f%%", per1 * 100),
		"per2": fmt.Sprintf("%.1f%%", per2 * 100),
		"id": fmt.Sprintf("%d", p.Id),
	}
}

func (p *Prediction) WaitAndStopAccepting() {
	if !p.StartedAt.IsZero() {
		return
	}
	secondsLeft := (p.CreatedAt.Unix() + int64(p.StartDelaySeconds)) - time.Now().Unix()
	if secondsLeft < 0 {
		secondsLeft = 0
	}
	log.Printf("predicion %d waiting %d seconds", p.Id, secondsLeft)
	time.Sleep(time.Second * time.Duration(secondsLeft))
	log.Printf("stop accepting prediction: %d", p.Id)
	err := p.db.StopAccepting(p)
	p.StartedAt = time.Now()
	if err != nil {
		log.Printf("stop accepting error: %s", err)
	}
}

func (p *Prediction) DeleteBet(uid UID) error {
	bet, ok := p.Bets[uid]
	if !ok {
		return &NotFound{}
	}
	err := p.db.DeleteBet(p, uid)
	if err != nil {
		return err
	}
	delete(p.Bets, uid)
	if bet.OnFirstOption {
		p.Balance1 -= bet.Amount
	} else {
		p.Balance2 -= bet.Amount
	}
	return nil
}
