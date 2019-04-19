package main

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var dbMutex = new(sync.Mutex)
var currentDB *sqlx.DB

// InitDB Init db connection
func InitDB(dsn string) (*sqlx.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if currentDB != nil {
		return currentDB, nil
	}

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	currentDB = db
	return currentDB, nil
}

// DBConn get db connection
func DBConn() *sqlx.DB {
	return currentDB
}

type Animal struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Place string `db:"place"`
}

func GetAnimal(id int64) (animal *Animal, err error) {
	animal = &Animal{}
	err = DBConn().Get(animal, `
		SELECT id, name, place
		FROM animal
		WHERE id=?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return animal, nil
}

func CreateAnimal(name string, place string) (animal *Animal, err error) {
	if name == "" || place == "" {
		return nil, fmt.Errorf("invalid params")
	}
	result, err := DBConn().Exec(
		`INSERT INTO animal (name, place) VALUES (?, ?)`,
		name, place)
	if err != nil {
		return nil, err
	}
	_id, _ := result.LastInsertId()
	return GetAnimal(_id)
}

func (a *Animal) ChangePlace(place string) (err error) {
	if place == "" {
		return fmt.Errorf("invalid param")
	}
	_, err = DBConn().Exec(`UPDATE animal SET place=?`, place)
	if err != nil {
		return err
	}
	a.Place = place
	return nil
}
