package webui

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/kamaln7/karmabot/database"
)

// HomeHandler handles / URLs
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/leaderboard", 302)
}

type leaderboardPageData struct {
	Limit, TotalPoints int
	Leaderboard        database.Leaderboard
}

// LeaderboardHandler handles /leaderboard URLs
func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	var (
		limit int
		err   error
	)
	limitS := mux.Vars(r)["limit"]

	if limitS == "" {
		limit = config.LeaderboardLimit
	} else {
		limit, err = strconv.Atoi(limitS)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	points, err := database.GetTotalPoints()

	leaderboard, err := database.GetLeaderboard(limit)
	if err != nil {
		config.Logger.Printf("error while generating the leaderbaord (limit %d): %s\n", limit, err.Error())

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := &templateData{
		Config: config,
		Data: leaderboardPageData{
			Limit:       limit,
			TotalPoints: points,
			Leaderboard: leaderboard,
		},
	}

	renderTemplate(w, "leaderboard.html", data)
}
