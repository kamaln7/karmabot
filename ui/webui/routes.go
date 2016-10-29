package webui

import (
	"net/http"
	"path"
)

func (u *UI) setupRoutes() {
	u.handlers = &Handlers{
		ui: u,
	}

	var (
		r = u.router
		h = u.handlers
	)

	// assets
	assetsPath := path.Join(u.Config.FilesPath, "assets")
	assetsHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsPath)))
	r.PathPrefix("/assets/").Handler(assetsHandler)

	// routes
	r.HandleFunc("/", h.MustAuth(h.Home)).Methods("GET")
	r.HandleFunc("/leaderboard", h.MustAuth(h.Leaderboard)).Methods("GET")
	r.HandleFunc(`/leaderboard/{limit:\d+}`, h.MustAuth(h.Leaderboard)).Methods("GET")

	// custom handlers
	r.NotFoundHandler = http.HandlerFunc(h.NotFound)
}
