package webui

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/kamaln7/karmabot/database"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/leaderboard", 302)
}

func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	var (
		limit int
		err   error
	)
	limitS := mux.Vars(r)["limit"]

	if limitS == "" {
		limit = leaderboardLimit
	} else {
		limit, err = strconv.Atoi(limitS)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte(fmt.Sprintf("Leaderboard limit at %d!\n", limit)))
}
