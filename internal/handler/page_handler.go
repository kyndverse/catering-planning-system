package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/internal/view"
	"golang-web/web"
)

type PageHandler struct {
	Renderer          *web.Renderer
	IngredientService *service.IngredientService
	RecipeService     *service.RecipesService
}

func NewPageHandler(renderer *web.Renderer, ingredientService *service.IngredientService, recipeService *service.RecipesService) *PageHandler {
	return &PageHandler{
		Renderer:          renderer,
		IngredientService: ingredientService,
		RecipeService:     recipeService,
	}
}

func (h *PageHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"title": "Home",
	}

	h.Renderer.Render(w, "home/index.html", data)
}

func (h *PageHandler) CreateIngredientForm(w http.ResponseWriter, r *http.Request) {
	ingredientForm := model.FormData[model.CreateIngredientRequest]{
		Judul:  "Create Ingredient",
		Action: "/ingredients",
		Data:   model.CreateIngredientRequest{},
		Errors: nil,
	}

	data := map[string]any{
		"title":          "Create Ingredient",
		"baseUnits":      []string{"g", "ml", "pcs"},
		"submitLabel":    "Simpan",
		"isEdit":         false,
		"IngredientForm": ingredientForm,
	}

	h.Renderer.Render(w, "ingredients/form.html", data)
}

func (h *PageHandler) EditIngredientForm(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	ingredient, err := h.IngredientService.GetIngredientById(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch ingredient", http.StatusInternalServerError)
		return
	}

	ingredientForm := model.FormData[model.UpdateIngredientRequest]{
		Judul:  "Update Ingredient",
		Action: fmt.Sprintf("/ingredients/%d/edit", ID),
		Data: model.UpdateIngredientRequest{
			Name:                   ingredient.Name,
			BaseUnit:               ingredient.BaseUnit,
			PurchaseUnit:           ingredient.PurchaseUnit,
			BaseQtyPerPurchaseUnit: ingredient.BaseQtyPerPurchaseUnit,
		},
		Errors: nil,
	}

	data := map[string]any{
		"title":          "Update Ingredient",
		"baseUnits":      []string{"g", "ml", "pcs"},
		"submitLabel":    "Simpan",
		"isEdit":         true,
		"IngredientForm": ingredientForm,
	}

	h.Renderer.Render(w, "ingredients/form.html", data)
}

func (h *PageHandler) IngredientPricePage(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	ing, err := h.IngredientService.GetIngredientById(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch ingredient", http.StatusInternalServerError)
		return
	}

	prices, err := h.IngredientService.GetAllPricesByIngredientID(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch ingredient", http.StatusInternalServerError)
		return
	}

	records := make([]view.PriceRecord, 0, len(prices))
	for _, p := range prices {
		records = append(records, view.PriceRecord{
			ID:                   int(p.ID),
			PricePerPurchaseUnit: fmt.Sprint(p.PricePerPurchaseUnit),
			EffectiveFrom:        p.EffectiveFrom,
		})
	}

	rows, currentPrice, err := view.BuildPriceRows(records, fmt.Sprint(ing.BaseQtyPerPurchaseUnit), time.Now())
	if err != nil {
		http.Error(w, "failed to build prices", http.StatusInternalServerError)
		return
	}

	ingredientForm := model.FormData[model.CreateIngredientPriceRequest]{
		Judul:  "Ingredient Price History",
		Action: fmt.Sprintf("/ingredients/%d/prices", ID),
		Data:   model.CreateIngredientPriceRequest{},
		Errors: nil,
	}

	data := map[string]any{
		"title":          "Price History",
		"currentPrice":   currentPrice,
		"Ingredient":     ing,
		"Prices":         rows,
		"IngredientForm": ingredientForm,
	}

	h.Renderer.Render(w, "ingredients/price.html", data)
}

func (h *PageHandler) CreateRecipeForm(w http.ResponseWriter, r *http.Request) {

	recipeFrom := model.FormData[model.CreateRecipeRequest]{
		Judul:  "Create Recipe",
		Action: "/recipes",
		Data:   model.CreateRecipeRequest{},
		Errors: nil,
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
}

func (h *PageHandler) EditRecipeForm(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	recipe, err := h.RecipeService.GetRecipeById(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch recipe", http.StatusInternalServerError)
		return
	}

	recipeFrom := model.FormData[model.UpdateRecipeRequest]{
		Judul:  "Update Recipe",
		Action: fmt.Sprintf("/recipes/%d/edit", ID),
		Data: model.UpdateRecipeRequest{
			Name:      recipe.Name,
			Type:      recipe.Type,
			YieldQty:  recipe.YieldQty,
			YieldUnit: recipe.YieldUnit,
		},
		Errors: nil,
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
}
