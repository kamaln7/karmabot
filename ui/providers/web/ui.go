package web

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kamaln7/karmabot/ui/providers/web/auth"
)

type UI struct {
	Config *Config

	handlers      *Handlers
	router        *mux.Router
	templates     *template.Template
	authenticator *auth.Authenticator
}

func newUI(config *Config) *UI {
	ui := &UI{
		Config: config,
		router: mux.NewRouter(),
		authenticator: auth.New(&auth.Config{
			Token: config.TOTP,
			Log:   config.Log.KV("service", "auth"),
		}),
	}

	ui.Init()
	return ui
}

func (u *UI) Init() {
	u.setupTemplates()
	u.setupRoutes()
}

func (u *UI) Listen() {
	u.Config.Log.KV("address", u.Config.ListenAddr).Info("starting http server")
	err := http.ListenAndServe(u.Config.ListenAddr, u.router)

	if err != nil {
		u.Config.Log.Err(err).Fatal("could not start http server")
	}
}
