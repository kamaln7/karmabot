package webui

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kamaln7/karmabot/database"

	"github.com/gorilla/mux"
)

// Handlers contains all the http.HandlerFuncs
// that serve the web UI's routes.
type Handlers struct {
	ui *UI
}

// Home redirects to the leaderboard view.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/leaderboard", 302)
}

// Leaderboard serves the leaderboard view.
func (h *Handlers) Leaderboard(w http.ResponseWriter, r *http.Request) {
	var (
		limit int
		err   error
	)

	limitS := mux.Vars(r)["limit"]

	if limitS == "" {
		limit = h.ui.Config.LeaderboardLimit
	} else {
		limit, err = strconv.Atoi(limitS)

		if err != nil {
			h.ui.renderError(w, err)
			return
		}
	}

	points, err := h.ui.Config.DB.GetTotalPoints()

	leaderboard, err := h.ui.Config.DB.GetLeaderboard(limit)
	if err != nil {
		h.ui.Config.Log.Err(err).KV("limit", limit).Error("could not generate leaderboard")

		h.ui.renderError(w, err)
		return
	}

	data := &templateData{
		Config: &templateConfig{
			LeaderboardLimit: h.ui.Config.LeaderboardLimit,
		},
		Data: &struct {
			Limit, TotalPoints int
			Leaderboard        database.Leaderboard
		}{
			Limit:       limit,
			TotalPoints: points,
			Leaderboard: leaderboard,
		},
	}

	h.ui.renderTemplate(w, "leaderboard.html", data)
}

// NotFound handles invalid URIs that do not
// have a matching route.
func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) {
	h.ui.renderTemplate(w, "error.html", &templateData{
		Data: &struct {
			Error string
		}{
			Error: fmt.Sprintf("page [%s] not found", r.RequestURI),
		},
	})
}
