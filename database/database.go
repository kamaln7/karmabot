package database

import (
	"database/sql"
	"strings"

	"github.com/aybabtme/log"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Path string
	Log  *log.Log
}

type DB struct {
	Config *Config
	SQL    *sql.DB
}

type Points struct {
	From, To, Reason string
	Points           int
}

// The Leaderboard lists the top X users
type Leaderboard []*User

// User is an entry in the Leaderboard
type User struct {
	Name   string
	Points int
}

func New(config *Config) (*DB, error) {
	instance := &DB{
		Config: config,
	}

	err := instance.Init()

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (db *DB) Init() error {
	sqlite, err := sql.Open("sqlite3", db.Config.Path)

	if err != nil {
		return err
	}

	db.SQL = sqlite
	return db.createTable()
}

func (db *DB) createTable() error {
	schema := `
		create table if not exists karma (
		^id^ integer primary key,
		^from^ text not null,
		^to^ text not null,
		^points^ integer not null,
		^reason^ text,
		^timestamp^ text not null default (datetime('now'))
	)`
	schema = strings.Replace(schema, "^", "`", -1)

	_, err := db.SQL.Exec(schema)
	if err != nil {
		return err
	}

	indexes := "create index if not exists idx_to on karma(`to`);"

	_, err = db.SQL.Exec(indexes)
	if err != nil {
		return err
	}

	return nil
}

// InsertPoints inserts a karma operation into the database
func (db *DB) InsertPoints(points *Points) error {
	stmt, err := db.SQL.Prepare("insert into karma (`from`, `to`, `reason`, `points`) values(?, ?, ?, ?)")

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(points.From, points.To, points.Reason, points.Points)

	return err
}

// GetUser returns info about a user
func (db *DB) GetUser(name string) (*User, error) {
	stmt, err := db.SQL.Prepare("select sum(`points`) as points from karma where `to` = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	user := &User{
		Name: name,
	}
	err = stmt.QueryRow(user).Scan(&user.Points)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetLeaderboard returns the leaderboard with the top X users
func (db *DB) GetLeaderboard(limit int) (Leaderboard, error) {
	rows, err := db.SQL.Query("select `to`,sum(points) as points from karma group by `to` order by `points` desc limit ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard Leaderboard
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Name, &user.Points)

		if err != nil {
			return nil, err
		}

		leaderboard = append(leaderboard, user)
	}

	return leaderboard, nil
}

// GetTotalPoints returns the amount of points given or taken
// for all users
func (db *DB) GetTotalPoints() (int, error) {
	var res int
	err := db.SQL.QueryRow("select sum(abs(`points`)) from karma").Scan(&res)

	if err != nil {
		return 0, err
	}

	return res, nil
}
