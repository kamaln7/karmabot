package webui

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kamaln7/karmabot/ui/webui/auth"
)

// A UI is the part of the web UI that handles
// everything HTTP i.e. the actual web UI.
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

// Init initializes the web UI by parsing the HTML
// templates and setting up the HTTP routes.
func (u *UI) Init() {
	u.setupTemplates()
	u.setupRoutes()
}

// Listen starts the actual HTTP server.
func (u *UI) Listen() {
	u.Config.Log.KV("address", u.Config.ListenAddr).Info("starting http server")
	err := http.ListenAndServe(u.Config.ListenAddr, u.router)

	if err != nil {
		u.Config.Log.Err(err).Fatal("could not start http server")
	}
}
