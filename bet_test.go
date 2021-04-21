package bet

import (
	"reflect"
	"testing"
)

func assertBalancesEqual(p *Prediction, b1 uint64, b2 uint64, t *testing.T) {
	if p.Balance1 != b1 || p.Balance2 != b2 {
		t.Fatalf("balnce not equal %d != %d, %d != %d", p.Balance1, b1, p.Balance2, b2)
	}
}

func assertBetEqual(uid UID, amount uint64, onFirstOption bool, p *Prediction, t *testing.T) {
	bet, ok := p.Bets[uid]
	if !ok {
		t.Fatalf("bet not found: %d", uid)
	}
	if bet.Amount != amount {
		t.Fatalf("amount mismatch %d != %d", bet.Amount, amount)
	}
	if bet.OnFirstOption != onFirstOption {
		t.Fatalf("on first option mismatch. Expected %t", onFirstOption)
	}
}

func assertNoError(err error, t *testing.T) {
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func assertErrorInstance(err error, errorType error, t *testing.T) {
	expectedType := reflect.TypeOf(errorType)
	actualType := reflect.TypeOf(err)
	if !expectedType.AssignableTo(actualType) {
		t.Fatalf("error type mimatch. expected %s, actual: %s", expectedType, actualType)
	}
}

type testDB struct{}

func (d *testDB) CreatePrediction(prediction *Prediction) error {
	return nil
}

func (d *testDB) CreateBet(prediction *Prediction, bet *Bet) error {
	return nil
}

func (d *testDB) DeleteBet(prediction *Prediction, uid UID) error {
	return nil
}

func (d *testDB) ClosePrediction(prediction *Prediction, opt1Won bool) error {
	return nil
}

func TestAddDeleteBet(t *testing.T) {
	db := &testDB{}
	pred, err := CreatePrediction("pre1", "o1", "o2", db)
	if err != nil {
		t.Error(err)
	}
	assertBalancesEqual(pred, 0, 0, t)

	err = pred.AddBet(Bet{UserId: 42, OnFirstOption: false, Amount: 1000})
	assertNoError(err, t)
	assertBalancesEqual(pred, 0, 1000, t)
	assertBetEqual(42, 1000, false, pred, t)

	err = pred.AddBet(Bet{UserId: 43, OnFirstOption: false, Amount: 2000})
	assertNoError(err, t)
	assertBalancesEqual(pred, 0, 3000, t)
	assertBetEqual(43, 2000, false, pred, t)

	err = pred.AddBet(Bet{UserId: 42, OnFirstOption: false, Amount: 1000})
	assertErrorInstance(err, &Duplicate{}, t)

	err = pred.DeleteBet(43)
	assertNoError(err, t)
	assertBalancesEqual(pred, 0, 1000, t)

	err = pred.AddBet(Bet{UserId: 44, OnFirstOption: true, Amount: 200})
	assertNoError(err, t)
	assertBalancesEqual(pred, 200, 1000, t)
	assertBetEqual(44, 200, true, pred, t)

	err = pred.DeleteBet(43)
	assertErrorInstance(err, &NotFound{}, t)

}
