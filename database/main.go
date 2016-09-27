package database

import (
	"database/sql"
	"log"
	"strings"

	// sqlite
	_ "github.com/mattn/go-sqlite3"
)

// Points describes a karma operation
type Points struct {
	From, To, Reason string
	Points           int
}

// The Leaderboard lists the top X users
type Leaderboard []LeaderboardUser

// LeaderboardUser is an entry in the Leaderboard
type LeaderboardUser struct {
	User   string
	Points int
}

var (
	ll *log.Logger
	db *sql.DB
)

// Init an sqlite instance
func Init(logger *log.Logger, path string) {
	ll = logger

	database, err := sql.Open("sqlite3", path)
	if err != nil {
		ll.Fatalf("could not open sqlite database: %s\n", err.Error())
	}

	db = database
	CreateTable()
}

// CreateTable creates the sql table if it does not exist
func CreateTable() {
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

	_, err := db.Exec(schema)
	if err != nil {
		ll.Fatalf("could not initialize sqlite database: %s\n", err.Error())
	}

	indexes := "create index if not exists idx_to on karma(`to`);"

	_, err = db.Exec(indexes)
	if err != nil {
		ll.Fatalf("could not create indexes: %s\n", err)
	}
}

// InsertPoints inserts a karma operation into the database
func InsertPoints(points Points) error {
	stmt, err := db.Prepare("insert into karma (`from`, `to`, `reason`, `points`) values(?, ?, ?, ?)")

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(points.From, points.To, points.Reason, points.Points)

	return err
}

// GetPoints returns the total amount of points for a user
func GetPoints(user string) (int, error) {
	stmt, err := db.Prepare("select sum(`points`) as points from karma where `to` = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var points int
	err = stmt.QueryRow(user).Scan(&points)
	if err != nil {
		return 0, err
	}

	return points, nil
}

// GetLeaderboard returns the leaderboard with the top X users
func GetLeaderboard(limit int) (Leaderboard, error) {
	rows, err := db.Query("select `to`,sum(points) as points from karma group by `to` order by `points` desc limit ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard Leaderboard
	for rows.Next() {
		user := LeaderboardUser{}
		err := rows.Scan(&user.User, &user.Points)

		if err != nil {
			return nil, err
		}

		leaderboard = append(leaderboard, user)
	}

	return leaderboard, nil
}
