package webui

import (
	"errors"
	"net/http"
)

// MustAuth wraps an http.HandlerFunc and ensures that the
// user is authenticated before the said HandlerFunc is
// executed. The user is redirected to a "session expired"
// page if they are not authenticated.
func (h *Handlers) MustAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authed, err := h.ui.authenticator.Authenticate(w, r)
		if err != nil {
			h.ui.renderError(w, err)
		}

		if authed {
			next(w, r)
		} else {
			h.ui.renderError(w, errors.New(`your session has expired. Please type "karmabot web" and click on the generated url`))
		}
	}
}
