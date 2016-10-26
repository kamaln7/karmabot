package web

import (
	"fmt"
	"net/http"
)

type Handlers struct {
	ui *UI
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}

func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) {
	h.ui.renderTemplate(w, "error.html", &templateData{
		Data: struct {
			Error string
		}{
			Error: fmt.Sprintf("page [%s] not found", r.RequestURI),
		},
	})
}
