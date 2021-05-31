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

	Total uint64
	Balance1 uint64
	Balance2 uint64

	Coef1 float64
	Coef2 float64
	Per1 float32
	Per2 float32
	NumPeople1 int32
	NumPeople2 int32

	CreatedBy uint64
	CreatedAt time.Time
	StartedAt time.Time
	FinishedAt time.Time
	Opt1Won bool

	StartDelaySeconds uint16

	started bool
	db      DB
}

type Predictions struct {
	PredictionIdMap map[uint64]*Prediction
	Predictions     []*Prediction
}

func (s *Predictions) Add(p *Prediction) {
	s.Predictions = append(s.Predictions, p)
	s.PredictionIdMap[p.Id] = p
}

func (s *Predictions) Delete(p *Prediction) {
	delete(s.PredictionIdMap, p.Id)
	i := 0
	for s.Predictions[i].Id != p.Id {
		i++
	}
	for i < len(s.Predictions) - 1 {
		s.Predictions[i] = s.Predictions[i + 1]
		i++
	}
	s.Predictions = s.Predictions[:i]
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
	p.UpdateStats(bet)
	return nil
}

func (p *Prediction) UpdateStats(bet Bet) {
	p.Bets[bet.UserId] = bet
	if bet.OnFirstOption {
		p.Balance1 += bet.Amount
		p.NumPeople1 += 1
	} else {
		p.Balance2 += bet.Amount
		p.NumPeople2 += 1
	}
	p.Total += bet.Amount
	p.Coef1 = float64(p.Total) / float64(p.Balance1)
	p.Coef2 = float64(p.Total) / float64(p.Balance2)
	p.Per1 = float32(p.Balance1) / float32(p.Total)
	p.Per2 = float32(p.Balance2) / float32(p.Total)
}

func (p *Prediction) GetBetInfoArgs() map[string]string {
	return map[string]string{
		"amountOpt1": fmt.Sprintf("%d", p.Balance1),
		"amountOpt2": fmt.Sprintf("%d", p.Balance2),
		"ppl1": fmt.Sprintf("%d", p.NumPeople1),
		"ppl2": fmt.Sprintf("%d", p.NumPeople2),
		"coef1": fmt.Sprintf("1:%.2f", p.Coef1),
		"coef2": fmt.Sprintf("1:%.2f", p.Coef2),
		"per1": fmt.Sprintf("%.0f%%", p.Per1 * 100),
		"per2": fmt.Sprintf("%.0f%%", p.Per2 * 100),
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
