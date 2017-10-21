package database

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/aybabtme/log"

	// import the sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

// Config contains the necessary config options to
// connect to an sqlite3 database.
type Config struct {
	Path string
	Log  *log.Log
}

// A DB in an instance of a karmabot database.
type DB struct {
	Config *Config
	SQL    *sql.DB
}

// Points is a karma record containing info about
// a karma operation.
type Points struct {
	From, To, Reason string
	Points           int
}

// The Leaderboard lists the top X users.
type Leaderboard []*User

// A User is an entry in the Leaderboard.
type User struct {
	Name   string
	Points int
}

// ErrNoSuchUser is returned when a user lookup
// is performed on a non-existent user
var ErrNoSuchUser = errors.New("no such user")

// New returns a new instance of a karmabot database
// and initializes it
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

// Init initializes an sqlite3 database in order
// for karmabot to be able to use it
func (db *DB) Init() error {
	sqlite, err := sql.Open("sqlite3", db.Config.Path)

	if err != nil {
		return err
	}

	db.SQL = sqlite
	return db.createTable()
}

func (db *DB) createTable() error {
	schema := strings.Replace(
		`create table if not exists karma (
			^id^ integer primary key,
			^from^ text not null,
			^to^ text not null,
			^points^ integer not null,
			^reason^ text,
			^timestamp^ text not null default (datetime('now'))
		)`,
		"^", "`", -1)

	_, err := db.SQL.Exec(schema)
	if err != nil {
		return err
	}

	_, err = db.SQL.Exec("create index if not exists idx_to on karma(`to`);")
	if err != nil {
		return err
	}

	return nil
}

// InsertPoints inserts a Points object into the database.
func (db *DB) InsertPoints(points *Points) error {
	stmt, err := db.SQL.Prepare("insert into karma (`from`, `to`, `reason`, `points`) values(?, ?, ?, ?)")

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(points.From, points.To, points.Reason, points.Points)

	return err
}

// GetUser returns info about a user.
func (db *DB) GetUser(name string) (*User, error) {
	stmt, err := db.SQL.Prepare("select count(`to`) as `count` from karma where `to` = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := &User{
		Name: name,
	}

	var userExists int
	err = stmt.QueryRow(user.Name).Scan(&userExists)
	if err != nil {
		return nil, err
	}
	if userExists == 0 {
		return nil, ErrNoSuchUser
	}

	stmt, err = db.SQL.Prepare("select sum(`points`) as `points` from karma where `to` = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(user.Name).Scan(&user.Points)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetLeaderboard returns the leaderboard with the top X users.
func (db *DB) GetLeaderboard(limit int) (Leaderboard, error) {
	rows, err := db.SQL.Query("select `to`, sum(`points`) as `points` from karma group by `to` order by `points` desc limit ?", limit)
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
// for all users.
func (db *DB) GetTotalPoints() (int, error) {
	var res int
	err := db.SQL.QueryRow("select sum(abs(`points`)) from karma").Scan(&res)

	if err != nil {
		return 0, err
	}

	return res, nil
}

// GetThrowback returns a random karma operation on a specific user
func (db *DB) GetThrowback(user string) (*Points, error) {
	record := &Points{}
	err := db.SQL.QueryRow("select `from`, `to`, `reason`, `points` from karma where `to` = ? and `id` >= (abs(random()) % (select max(`id`) from karma)) limit 1", user).Scan(&record.From, &record.To, &record.Reason, &record.Points)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return record, nil
}
