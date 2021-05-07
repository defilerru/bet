package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type MySQLDB struct {
	db                   *sql.DB
	stmtCreatePrediction *sql.Stmt
	stmtSelectPrediction *sql.Stmt
	stmtGetGasInfo *sql.Stmt
}

const (
	stmtCreatePrediction = "INSERT INTO predictions(name, option_1, option_2, start_delay_seconds, created_by) VALUES(?,?,?,?,?)"
	stmtSelectPrediction = "SELECT created_at, started_at FROM predictions WHERE id = ?"
	stmtCreateBet        = "INSERT INTO bets(user_id, prediction_id, amount, on_first_option) VALUES(?,?,?,?)"
	stmtTakePayment      = "UPDATE bets_users SET balance=balance-? WHERE id=? AND balance>=?"
	stmtGetGasInfo       = "SELECT balance FROM bets_users WHERE id=?"
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

	mysqlDB.stmtGetGasInfo, err = db.Prepare(stmtGetGasInfo)
	if err != nil {
		return mysqlDB, err
	}

	return mysqlDB, err
}

func (m *MySQLDB) GetGasInfo(uid UID) (int64, error) {
	var gas int64
	err := m.stmtGetGasInfo.QueryRow(uid).Scan(&gas)
	return gas, err
}

func (m *MySQLDB) CreatePrediction(prediction *Prediction) error {
	res, err := m.stmtCreatePrediction.Exec(prediction.Name, prediction.Opt1, prediction.Opt2, prediction.StartDelaySeconds, prediction.CreatedBy)
	if err != nil {
		return err
	}
	rowID, err := res.LastInsertId()
	prediction.Id = uint64(rowID)
	return err
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
