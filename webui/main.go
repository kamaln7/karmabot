package webui

import (
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"
	//"github.com/gorilla/sessions"
)

// Config sets the different config options that are
// needed to start the webserver
type Config struct {
	Logger                         *log.Logger
	TOTPKey, ListenAddr, FilesPath string
	LeaderboardLimit               int
	Debug                          bool
}

var (
	config *Config
	router *mux.Router
)

// Init initiates the web ui config
func Init(c *Config) {
	config = c

	if config.TOTPKey == "" {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "karmabot",
			AccountName: "slack",
		})

		if err != nil {
			config.Logger.Fatalf("an error occurred while generating a TOTP key: %s\n", err.Error())
		} else {
			config.Logger.Fatalf("please use the following TOTP key (`karmabot -totp <key>`): %s\n", key.Secret())
		}
	}

	setupTemplates()

	router = mux.NewRouter()
	setupRoutes(router)
	go expireClients()
}

func setupRoutes(r *mux.Router) {
	r.Handle("/", mustAuth(HomeHandler))
	r.Handle("/leaderboard", mustAuth(LeaderboardHandler))
	r.Handle(`/leaderboard/{limit:\d+}`, mustAuth(LeaderboardHandler))

	assetsPath := path.Join(config.FilesPath, "assets")
	assetsHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsPath)))
	r.PathPrefix("/assets").Handler(assetsHandler)
}

// GetToken generates and returns a TOTP token
func GetToken() (string, error) {
	return totp.GenerateCode(config.TOTPKey, time.Now())
}

// Listen starts the web ui
func Listen() {
	config.Logger.Printf("serving webui on %s\n", config.ListenAddr)
	config.Logger.Fatal(http.ListenAndServe(config.ListenAddr, router))
}
