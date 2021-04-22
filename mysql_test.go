package bet

import (
	"flag"
	"testing"
)

var skipmysql = flag.Bool("skipmysql", false, "only perform local tests")

func assertBool(actual bool, expected bool, t *testing.T) {
	if actual != expected {
		t.Fatalf("bool fail %t != %t", actual, expected)
	}
}

func TestCreateClosePrediction(t *testing.T) {
	if *skipmysql {
		t.Skip("skipmysql flag is not set")
	}
	db, err := NewMySQLDB("defiler@/defiler_test?parseTime=true&loc=Local")
	if err != nil {
		t.Fatalf("can't connect to skipmysql: %s", err)
	}
	prediction, err := CreatePrediction("Test 1", "Yes", "No", 40, db)
	err = db.CreatePrediction(prediction)
	if err != nil {
		t.Fatalf("can't create prediction: %s", err)
	}

	err = prediction.AddBet(Bet{
		UserId:        41,
		Amount:        200,
		OnFirstOption: false,
	})
	assertNoError(err, t)

	err = prediction.AddBet(Bet{
		UserId:        41,
		Amount:        200,
		OnFirstOption: false,
	})
	assertErrorInstance(err, &Duplicate{}, t)

	err = prediction.AddBet(Bet{
		UserId:        42,
		Amount:        200000,
		OnFirstOption: false,
	})
	assertErrorInstance(err, &Insufficient{}, t)

	predictionLoaded, err := db.LoadPrediction(prediction.Id)
	if err != nil {
		t.Fatalf("can't load prediction: %s", err)
	}
	assertBool(predictionLoaded.CreatedAt.IsZero(), false, t)
	assertBool(predictionLoaded.StartedAt.IsZero(), true, t)

	err = db.ClosePrediction(prediction, true)
	assertNoError(err, t)
}
