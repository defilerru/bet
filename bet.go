package main

import (
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
