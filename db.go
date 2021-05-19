package main

type DB interface {
	CreatePrediction(prediction *Prediction) error
	LoadPredictions() (map[uint64]*Prediction, error)
	CreateBet(prediction *Prediction, bet *Bet) error
	DeleteBet(prediction *Prediction, uid UID) error
	ClosePrediction(prediction *Prediction, opt1Won bool) error
	GetUserInfo(UID, *int64, *string, *bool) error
	StopAccepting(prediction *Prediction) error
}
