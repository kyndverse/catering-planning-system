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

// dashboard soon
func (h *PageHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/ingredients", http.StatusSeeOther)
}
