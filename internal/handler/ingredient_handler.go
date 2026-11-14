package handler

import (
	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/web"
	"net/http"

	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
)

type IngredientHandler struct {
	Log               *logrus.Logger
	Renderer          *web.Renderer
	Decoder           *form.Decoder
	IngredientService *service.IngredientService
}

func NewIngredientHandler(logger *logrus.Logger, renderer *web.Renderer, decoder *form.Decoder, ingredientService *service.IngredientService) *IngredientHandler {
	return &IngredientHandler{
		Log:               logger,
		Renderer:          renderer,
		Decoder:           decoder,
		IngredientService: ingredientService,
	}
}

func (h *IngredientHandler) GetAllIngredient(w http.ResponseWriter, r *http.Request) {
	ingredients, err := h.IngredientService.GetAllIngredients(r.Context())
	if err != nil {
		h.Log.WithError(err).Error("error get ingredients")
		http.Error(w, "Failed to retrieve ingredients", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"title":       "Ingredients",
		"ingredients": ingredients,
	}

	h.Renderer.Render(w, "ingredients-page.html", data)
}

func (h *IngredientHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.CreateIngredientRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	_, formError, err := h.IngredientService.AddIngredient(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error add ingredients")
		http.Error(w, "Failed to add ingredients", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		ingredientForm := model.FormData[model.CreateIngredientRequest]{
			Judul:  "Create Product",
			Action: "/ingredients",
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":          "Create Product",
			"baseUnits":      []string{"g", "ml", "pcs"},
			"submitLabel":    "Simpan",
			"isEdit":         false,
			"IngredientForm": ingredientForm,
		}

		h.Renderer.Render(w, "ingredients-form-page.html", data)
		return
	}

	http.Redirect(w, r, "/ingredients", http.StatusSeeOther)
}
