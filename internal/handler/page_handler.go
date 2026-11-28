package handler

import (
	"net/http"

	"golang-web/web"
)

type PageHandler struct {
	Renderer *web.Renderer
}

func NewPageHandler(renderer *web.Renderer) *PageHandler {
	return &PageHandler{
		Renderer: renderer,
	}
}

func (h *PageHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"title": "Home",
	}

	h.Renderer.Render(w, "home/index.html", data)
}
