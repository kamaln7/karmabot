package webui

import (
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type templateConfig struct {
	LeaderboardLimit int
}

type templateData struct {
	Config *templateConfig
	Data   interface{}
}

func (u *UI) setupTemplates() {
	u.templates = template.New("")

	templatesPath := path.Join(u.Config.FilesPath, "templates")
	err := filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			_, err = u.templates.ParseFiles(path)

			if err != nil {
				u.Config.Log.Err(err).KV("template", path).Error("could not parse template")
			}
		}

		return err
	})

	if err != nil {
		u.Config.Log.Err(err).Fatal("could not parse all templates properly. exiting.")
	}
}

func (u *UI) renderTemplate(w http.ResponseWriter, tmpl string, data *templateData) {
	err := u.templates.ExecuteTemplate(w, tmpl, data)

	if err != nil {
		u.Config.Log.Err(err).KV("template", tmpl).Error("could not render template")

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (u *UI) renderError(w http.ResponseWriter, err error) {
	u.renderTemplate(w, "error.html", &templateData{
		Config: nil,
		Data: &struct {
			Error string
		}{
			Error: err.Error(),
		},
	})
}
