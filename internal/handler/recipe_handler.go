package handler

import (
	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/web"
	"net/http"

	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
)

type RecipeHandler struct {
	Log           *logrus.Logger
	Renderer      *web.Renderer
	Decoder       *form.Decoder
	RecipeService *service.RecipesService
}

func NewRecipesHandler(logger *logrus.Logger, renderer *web.Renderer, decoder *form.Decoder, recipeService *service.RecipesService) *RecipeHandler {
	return &RecipeHandler{
		Log:           logger,
		Renderer:      renderer,
		Decoder:       decoder,
		RecipeService: recipeService,
	}
}

func (h *RecipeHandler) Search(w http.ResponseWriter, r *http.Request) {
	typeQuery := r.URL.Query().Get("type")

	request := &model.SearchRecipeRequest{
		Type: typeQuery,
	}

	recipes, err := h.RecipeService.Search(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error get recipes")
		http.Error(w, "Failed to retrieve recipes", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"title":      "Recipes List",
		"filterType": typeQuery,
		"recipes":    recipes,
	}

	h.Renderer.Render(w, "recipes/index.html", data)
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.CreateRecipeRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	_, formError, err := h.RecipeService.AddRecipe(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error add recipe")
		http.Error(w, "Failed to add recipe", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		recipeFrom := model.FormData[model.CreateRecipeRequest]{
			Judul:  "Create Recipe",
			Action: "/recipes",
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":       "Create Recipe",
			"isEdit":      false,
			"submitLabel": "Create",
			"cancelURL":   "/recipes",
			"types":       []string{"PREPARED", "COMPONENT", "MENU"},
			"units":       []string{"g", "ml", "portion"},
			"RecipeForm":  recipeFrom,
		}

		h.Renderer.Render(w, "recipes/form.html", data)
		return
	}

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}
