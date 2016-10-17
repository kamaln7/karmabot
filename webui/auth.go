package webui

import (
	"net/http"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/satori/go.uuid"
)

type authMiddleware struct {
	handler http.Handler
}

type Client struct {
	UUID  string
	Added time.Time
}

var (
	authedClients []*Client
	clientsMutex  sync.RWMutex
)

func mustAuth(h http.HandlerFunc) *authMiddleware {
	return &authMiddleware{http.HandlerFunc(h)}
}

func (h *authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")

	if err != nil {
		if err != http.ErrNoCookie {
			config.Logger.Printf("could not auth user: %s\n", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		err = nil
	}

	authed := false
	if cookie != nil {
		clientsMutex.RLock()
		for _, client := range authedClients {
			if cookie.Value == client.UUID {
				authed = true
				break
			}
		}
		clientsMutex.RUnlock()
	}

	if !authed && hasValidTOTPToken(r) {
		cookie = &http.Cookie{
			Name:  "session",
			Value: uuid.NewV4().String(),
		}

		clientsMutex.Lock()
		authedClients = append(authedClients, &Client{
			UUID:  cookie.Value,
			Added: time.Now(),
		})
		clientsMutex.Unlock()

		authed = true
		http.SetCookie(w, cookie)
	}

	if !authed {
		renderTemplate(w, "unauthed.html", &templateData{
			Config: config,
			Data:   nil,
		})
		return
	}

	h.handler.ServeHTTP(w, r)
}

func hasValidTOTPToken(r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token == "" {
		return false
	}

	return totp.Validate(token, config.TOTPKey)
}

func expireClients() {
	for {
		<-time.After(2 * time.Minute)
		go func() {
			now := time.Now()

			for i, client := range authedClients {
				if now.Sub(client.Added).Hours() >= 48 {
					authedClients = append(authedClients[:i], authedClients[i+1:]...)
				}
			}
		}()
	}
}
