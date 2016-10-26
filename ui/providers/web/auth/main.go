package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/aybabtme/log"
	"github.com/pquerna/otp/totp"
	"github.com/satori/go.uuid"
)

type Config struct {
	Token string
	Log   *log.Log
}

type Authenticator struct {
	Config        *Config
	authedClients []*Client
	clientsMutex  sync.RWMutex
}

// Client is an authenticated web ui session
type Client struct {
	UUID  string
	Added time.Time
}

func New(config *Config) *Authenticator {
	authenticator := &Authenticator{
		Config: config,
	}

	go authenticator.ExpireClients()
	return authenticator
}

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

func (a *Authenticator) ExpireClients() {
	for {
		<-time.After(2 * time.Minute)
		go func() {
			now := time.Now()

			for i, client := range a.authedClients {
				if now.Sub(client.Added).Hours() >= 48 {
					a.authedClients = append(a.authedClients[:i], a.authedClients[i+1:]...)
				}
			}
		}()
	}
}

func (a *Authenticator) GetToken() (string, error) {
	return totp.GenerateCode(a.Config.Token, time.Now())
}
