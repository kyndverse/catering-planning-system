package handler

import (
	"net/http"

	"golang-web/internal/model"
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
	h.Renderer.Render(w, "home-page.html", nil)
}

func (h *PageHandler) CreateIngredientForm(w http.ResponseWriter, r *http.Request) {
	ingredientForm := model.FormData[model.CreateIngredientRequest]{
		Judul:  "Add Ingredients",
		Action: "/ingredients",
		Data:   model.CreateIngredientRequest{},
		Errors: nil,
	}

	data := map[string]any{
		"title":          "Create Product",
		"baseUnits":      []string{"g", "ml", "pcs"},
		"submitLabel":    "Simpan",
		"isEdit":         false,
		"IngredientForm": ingredientForm,
	}

	h.Renderer.Render(w, "ingredients-form-page.html", data)
}
