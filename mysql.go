package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type MySQLDB struct {
	db                    *sql.DB
	stmtCreatePrediction  *sql.Stmt
	stmtSelectPrediction  *sql.Stmt
	stmtGetUserInfo       *sql.Stmt
	stmtSelectPredictions *sql.Stmt
}

const (
	stmtCreatePrediction  = "INSERT INTO predictions(name, option_1, option_2, start_delay_seconds, created_by, created_at) VALUES(?,?,?,?,?,?)"
	stmtSelectPrediction  = "SELECT created_at, started_at FROM predictions WHERE id = ?"
	stmtCreateBet         = "INSERT INTO bets(user_id, prediction_id, amount, on_first_option) VALUES(?,?,?,?)"
	stmtTakePayment       = "UPDATE bets_users SET balance=balance-? WHERE id=? AND balance>=?"
	stmtGetUserInfo       = "SELECT username, balance, moderator FROM bets_users WHERE id=?"
	stmtSelectPredictions = "SELECT id, created_by, created_at, started_at, name, option_1, option_2, opt1_won, start_delay_seconds FROM predictions WHERE finished_at IS NULL ORDER BY created_at"
)

func NewMySQLDB(dataSourceName string) (*MySQLDB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	mysqlDB := &MySQLDB{
		db: db,
	}
	if err != nil {
		return mysqlDB, err
	}

	mysqlDB.stmtCreatePrediction, err = db.Prepare(stmtCreatePrediction)
	if err != nil {
		return mysqlDB, err
	}

	mysqlDB.stmtSelectPrediction, err = db.Prepare(stmtSelectPrediction)
	if err != nil {
		return mysqlDB, err
	}

	mysqlDB.stmtGetUserInfo, err = db.Prepare(stmtGetUserInfo)
	if err != nil {
		return mysqlDB, err
	}

	mysqlDB.stmtSelectPredictions, err = db.Prepare(stmtSelectPredictions)
	if err != nil {
		return mysqlDB, err
	}

	return mysqlDB, err
}

func (m *MySQLDB) GetUserInfo(uid UID, gas *int64, username *string, moderator *bool) error {
	err := m.stmtGetUserInfo.QueryRow(uid).Scan(username, gas, moderator)
	return err
}

func (m *MySQLDB) CreatePrediction(prediction *Prediction) error {
	prediction.CreatedAt = time.Now()
	res, err := m.stmtCreatePrediction.Exec(prediction.Name,
		prediction.Opt1,
		prediction.Opt2,
		prediction.StartDelaySeconds,
		prediction.CreatedBy,
		prediction.CreatedAt)
	if err != nil {
		return err
	}
	rowID, err := res.LastInsertId()
	prediction.Id = uint64(rowID)
	return err
}

func (m *MySQLDB) LoadPredictions() (preds map[uint64]*Prediction, err error) {
	preds = map[uint64]*Prediction{}
	rows, err := m.stmtSelectPredictions.Query()
	if err != nil {
		return
	}
	for rows.Next() {
		startedAt := sql.NullTime{}
		pred := &Prediction{}
		err = rows.Scan(
			&pred.Id,
			&pred.CreatedBy,
			&pred.CreatedAt,
			&startedAt,
			&pred.Name,
			&pred.Opt1,
			&pred.Opt2,
			&pred.Opt1Won,
			&pred.StartDelaySeconds)
		if err != nil {
			return
		}
		pred.StartedAt = startedAt.Time
		//TODO: load bets
		preds[pred.Id] = pred
	}
	return
}

func (m *MySQLDB) LoadPrediction(id uint64) (*Prediction, error) {
	p := &Prediction{}
	nt := sql.NullTime{}
	err := m.stmtSelectPrediction.QueryRow(id).Scan(&p.CreatedAt, &nt)
	p.StartedAt = nt.Time
	//TODO: the rest
	return p, err
}

func (m *MySQLDB) ClosePrediction(prediction *Prediction, opt1Won bool) error {
	tx, err := m.db.BeginTx(context.TODO(), nil)
	if err != nil {
		return err
	}
	for _, bet := range prediction.Bets {
		log.Println(bet)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLDB) CreateBet(prediction *Prediction, bet *Bet) error {
	tx, err := m.db.BeginTx(context.TODO(), nil)
	if err != nil {
		return err
	}
	stmtPay, err := tx.Prepare(stmtTakePayment)
	if err != nil {
		return err
	}
	res, err := stmtPay.Exec(bet.Amount, bet.UserId, bet.Amount)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return &Insufficient{}
	}
	stmt, err := tx.Prepare(stmtCreateBet)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(bet.UserId, prediction.Id, bet.Amount, bet.OnFirstOption)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (m *MySQLDB) DeleteBet(prediction *Prediction, uid UID) error {
	return nil
}
