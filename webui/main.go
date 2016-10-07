package webui

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"
	//"github.com/gorilla/sessions"
)

var (
	ll                             *log.Logger
	totpKey, listenAddr, filesPath string
	router                         *mux.Router
	leaderboardLimit               int
)

// Init initiates the web ui config
func Init(logger *log.Logger, key, addr, path string, limit int) {
	ll = logger
	totpKey = key
	listenAddr = addr
	filesPath = path
	leaderboardLimit = limit

	if totpKey == "" {
		newKey, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "karmabot",
			AccountName: "slack",
		})

		if err != nil {
			ll.Fatalf("an error occurred while generating a TOTP key: %s\n", err.Error())
		} else {
			ll.Fatalf("please use the following TOTP key (`karmabot -totp <key>`): %s\n", newKey.Secret())
		}
	}

	router = mux.NewRouter()
	setupRoutes(router)
}

func setupRoutes(r *mux.Router) {
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/leaderboard", LeaderboardHandler)
	r.HandleFunc(`/leaderboard/{limit:\d+}`, LeaderboardHandler)
}

// GetToken generates and returns a TOTP token
func GetToken() (string, error) {
	return totp.GenerateCode(totpKey, time.Now())
}

// Listen starts the web ui
func Listen() {
	ll.Printf("serving webui on %s\n", listenAddr)
	ll.Fatal(http.ListenAndServe(listenAddr, router))
}
