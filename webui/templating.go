package webui

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type templateData struct {
	Config *Config
	Data   interface{}
}

var templates *template.Template

func setupTemplates() {
	templ := template.New("")

	err := filepath.Walk(config.FilesPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			_, err = templ.ParseFiles(path)

			if err != nil {
				config.Logger.Printf("could not parse template [%s]: %s\n", path, err.Error())
			}
		}

		return err
	})

	if err != nil {
		config.Logger.Fatalln("could not parse all templates properly. exiting.")
	}

	templates = templ
}

func renderTemplate(w http.ResponseWriter, tmpl string, data *templateData) {
	err := templates.ExecuteTemplate(w, tmpl, *data)

	if err != nil {
		config.Logger.Printf("could not render template [%s]: %s\n", tmpl, err.Error())

		out := http.StatusText(http.StatusInternalServerError)
		if config.Debug {
			out = err.Error()
		}

		http.Error(w, out, http.StatusInternalServerError)
	}
}
