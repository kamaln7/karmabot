package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/aybabtme/log"
	"github.com/pquerna/otp/totp"
	"github.com/satori/go.uuid"
)

// Config contains the config options for the
// TOTP authentication serivce that is used
// for the web UI.
type Config struct {
	Token string
	Log   *log.Log
}

// An Authenticator contains a list of authenticated
// web UI sessions and exposes a few functions
// for authenticating users and generating tokens.
type Authenticator struct {
	Config        *Config
	authedClients []*Client
	clientsMutex  sync.RWMutex
}

// A Client is an authenticated web UI session
type Client struct {
	UUID  string
	Added time.Time
}

// New returns a new Authenticator instance and spins
// up a goroutine that handles expiring sessions
func New(config *Config) *Authenticator {
	authenticator := &Authenticator{
		Config: config,
	}

	go authenticator.ExpireClients()
	return authenticator
}

// Authenticate logs in the client if the request contains
// a token and checks whether the current request is
// authenticated.
func (a *Authenticator) Authenticate(w http.ResponseWriter, r *http.Request) (bool, error) {
	cookie, err := r.Cookie("session")

	if err != nil {
		if err != http.ErrNoCookie {
			a.Config.Log.Err(err).Error("could not authenticate user")

			return false, err
		}

		err = nil
	}

	authed := false
	if cookie != nil {
		a.clientsMutex.RLock()
		for _, client := range a.authedClients {
			if cookie.Value == client.UUID {
				authed = true
				break
			}
		}
		a.clientsMutex.RUnlock()
	}

	if !authed && a.hasValidToken(r) {
		cookie = &http.Cookie{
			Name:  "session",
			Value: uuid.NewV4().String(),
			Path:  "/",
		}

		a.clientsMutex.Lock()
		a.authedClients = append(a.authedClients, &Client{
			UUID:  cookie.Value,
			Added: time.Now(),
		})
		a.clientsMutex.Unlock()

		authed = true
		http.SetCookie(w, cookie)
	}

	return authed, nil
}

func (a *Authenticator) hasValidToken(r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token == "" {
		return false
	}

	return totp.Validate(token, a.Config.Token)
}

// ExpireClients expires all sessions that have been logged
// in for 48+ hours.
func (a *Authenticator) ExpireClients() {
	for {
		<-time.After(2 * time.Minute)

		now := time.Now()

		a.clientsMutex.Lock()
		for i := len(a.authedClients) - 1; i >= 0; i-- {
			if now.Sub(a.authedClients[i].Added).Hours() >= 48 {
				a.authedClients = append(a.authedClients[:i], a.authedClients[i+1:]...)
			}
		}
		a.clientsMutex.Unlock()
	}
}

// GetToken generates a TOTP token that is
// valid for 30 seconds.
func (a *Authenticator) GetToken() (string, error) {
	return totp.GenerateCode(a.Config.Token, time.Now())
}
