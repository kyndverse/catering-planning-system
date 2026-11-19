package handler

import (
	"fmt"
	"golang-web/internal/handler/view"
	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/web"
	"net/http"
	"strconv"

	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
)

type RecipeHandler struct {
	Log               *logrus.Logger
	Renderer          *web.Renderer
	Decoder           *form.Decoder
	RecipeService     *service.RecipesService
	IngredientService *service.IngredientService
}

func NewRecipesHandler(logger *logrus.Logger, renderer *web.Renderer, decoder *form.Decoder, recipeService *service.RecipesService, ingredientService *service.IngredientService) *RecipeHandler {
	return &RecipeHandler{
		Log:               logger,
		Renderer:          renderer,
		Decoder:           decoder,
		RecipeService:     recipeService,
		IngredientService: ingredientService,
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

func (h *RecipeHandler) EditRecipe(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateRecipeRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	request.ID = ID
	_, formError, err := h.RecipeService.EditRecipe(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error update recipe")
		http.Error(w, "Failed to update recipe", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		recipeFrom := model.FormData[model.UpdateRecipeRequest]{
			Judul:  "Update Recipe",
			Action: fmt.Sprintf("/ingredients/%d/edit", ID),
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":       "Update Recipe",
			"isEdit":      true,
			"submitLabel": "Update",
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

func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	request := new(model.DeleteRecipeRequest)
	request.ID = ID

	if err := h.RecipeService.DeleteRecipe(r.Context(), request); err != nil {
		h.Log.WithError(err).Error("error delete recipe")
		http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

func (h *RecipeHandler) GetRecipeDetail(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	recipe, err := h.RecipeService.GetRecipeById(r.Context(), ID)
	if err != nil {
		http.Error(w, "Failed to retrieve recipe", http.StatusInternalServerError)
		return
	}

	items, err := h.RecipeService.ListItemsByRecipeID(r.Context(), ID)
	if err != nil {
		http.Error(w, "Failed to retrieve recipe items", http.StatusInternalServerError)
		return
	}

	itemForm := model.FormData[model.AddRecipeItemRequest]{
		Action: fmt.Sprintf("/recipes/%d/items", ID),
		Data:   model.AddRecipeItemRequest{},
		Errors: nil,
	}

	ingredients, err := h.IngredientService.GetAllIngredients(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve ingredients", http.StatusInternalServerError)
		return
	}

	subRecipeOptions, err := h.RecipeService.ListSubRecipeOptions(r.Context(), ID)
	if err != nil {
		http.Error(w, "Failed to retrieve ingredients", http.StatusInternalServerError)
		return
	}

	qty, qtyOK := view.ResolveExplosionQty(
		r.URL.Query().Get("qty"),
		strconv.FormatFloat(recipe.YieldQty, 'f', -1, 64),
		recipe.YieldUnit,
	)

	reqs, err := h.RecipeService.ExplodeToIngredients(r.Context(), ID, qty)
	if err != nil {
		h.Log.WithError(err).Error("error explode recipe")
		http.Error(w, "Failed to calculate ingredient requirements", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"title":             "Recipe Detail",
		"active":            "recipes",
		"recipe":            recipe,
		"items":             items,
		"itemForm":          itemForm,
		"ingredientOptions": ingredients,
		"subRecipeOptions":  subRecipeOptions,
		"explosion":         view.NewExplosion(qty, reqs),
		"qtyInvalid":        !qtyOK,
	}

	h.Renderer.Render(w, "recipes/show.html", data)
}
