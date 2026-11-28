package web

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"text/template"
)

//go:embed templates
var templateFS embed.FS

var funcs = template.FuncMap{}

type Renderer struct {
	template map[string]*template.Template
}

func NewRenderer() *Renderer {
	renderer := &Renderer{
		template: make(map[string]*template.Template),
	}

	baseFiles := []string{
		"templates/layout.html",
		"templates/sidebar.html",
		"templates/partials/form-errors.html",
	}

	pages, err := fs.Glob(templateFS, "templates/pages/*/*.html")
	if err != nil {
		panic(err)
	}

	for _, page := range pages {
		name := strings.TrimPrefix(page, "templates/pages/")
		files := append(append([]string{}, baseFiles...), page)
		tmpl, err := template.New("layout").Funcs(funcs).ParseFS(templateFS, files...)
		if err != nil {
			panic(fmt.Sprintf("parse template %q: %v", page, err))
		}

		renderer.template[name] = tmpl
	}

	return renderer
}

func (r *Renderer) Render(w http.ResponseWriter, file string, data any) {
	tmpl, ok := r.template[file]
	if !ok {
		log.Printf("render: template %q not found", file)
		http.Error(w, "Halaman tidak ditemukan", http.StatusNotFound)
		return
	}

	var buffer bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buffer, "layout", data); err != nil {
		log.Printf("render %s: %v", file, err)
		http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buffer.WriteTo(w)
}
