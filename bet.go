package bet

type UID uint64

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
	Id      uint64
	Name    string
	Opt1    string
	Opt2    string
	Bets    map[UID]Bet
	Balance uint64

	db DB
}

type Bet struct {
	UserId        UID
	Amount        uint64
	OnFirstOption bool
}

func CreatePrediction(name string, opt1 string, opt2 string, db DB) (prediction *Prediction, err error) {
	prediction = &Prediction{
		Name: name,
		Opt1: opt1,
		Opt2: opt2,
		Bets: map[UID]Bet{},
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
	p.Balance += bet.Amount
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
	p.Balance -= bet.Amount
	return nil
}
