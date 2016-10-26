package web

import (
	"errors"
	"net/http"
)

func (h *Handlers) MustAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authed, err := h.ui.authenticator.Authenticate(w, r)
		if err != nil {
			h.ui.renderError(w, err)
		}

		if authed {
			next(w, r)
		} else {
			h.ui.renderError(w, errors.New(`Your session has expired. Please type "karmabot web" and click on the generated url.`))
		}
	}
}
