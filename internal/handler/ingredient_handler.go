package handler

import (
	"fmt"
	"golang-web/internal/handler/view"
	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/web"
	"net/http"
	"strconv"
	"time"

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
			Judul:  "Create Ingredient",
			Action: "/ingredients",
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":          "Create Ingredient",
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

func (h *IngredientHandler) EditIngredient(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	h.Log.Info("ID params adalah: ", ID)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateIngredientRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	request.ID = ID
	_, formError, err := h.IngredientService.EditIngredient(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error edit ingredients")
		http.Error(w, "Failed to edit ingredients", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		ingredientForm := model.FormData[model.UpdateIngredientRequest]{
			Judul:  "Update Ingredient",
			Action: fmt.Sprintf("/ingredients/%d/edit", ID),
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":          "Update Ingredient",
			"baseUnits":      []string{"g", "ml", "pcs"},
			"submitLabel":    "Simpan",
			"isEdit":         true,
			"IngredientForm": ingredientForm,
		}

		h.Renderer.Render(w, "ingredients-form-page.html", data)
		return
	}

	http.Redirect(w, r, "/ingredients", http.StatusSeeOther)
}

func (h *IngredientHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	request := new(model.DeleteIngredientRequest)
	request.ID = ID

	if err := h.IngredientService.DeleteIngredient(r.Context(), request); err != nil {
		h.Log.WithError(err).Error("error delete ingredients")
		http.Error(w, "Failed to delete ingredients", http.StatusInternalServerError)
		return
	}

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

func (h *IngredientHandler) AddIngredientPrice(w http.ResponseWriter, r *http.Request) {
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

	request := new(model.CreateIngredientPriceRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		h.Log.WithError(err).Error("Error decode form")
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	effectiveFromDate := r.PostFormValue("effective_from")
	effectiveFromDateOnly, err := time.Parse(time.DateOnly, effectiveFromDate)
	if err != nil {
		http.Error(w, "failed to add ingredient price", http.StatusInternalServerError)
		return
	}

	request.IngredientID = ID
	request.EffectiveFrom = effectiveFromDateOnly

	_, formError, err := h.IngredientService.AddIngredientPrice(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error add ingredient price")
		http.Error(w, "Failed to add ingredient price", http.StatusInternalServerError)
		return
	}

	ing, err := h.IngredientService.GetIngredientById(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch ingredient", http.StatusInternalServerError)
		return
	}

	prices, err := h.IngredientService.GetAllPricesByIngredientID(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch ingredient prices", http.StatusInternalServerError)
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

	if formError != nil {
		ingredientForm := model.FormData[model.CreateIngredientPriceRequest]{
			Judul:  "Create Ingredient Price",
			Action: fmt.Sprintf("/ingredients/%d/prices", ID),
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":          "Edit Product",
			"currentPrice":   currentPrice,
			"Ingredient":     ing,
			"Prices":         rows,
			"IngredientForm": ingredientForm,
		}
		h.Renderer.Render(w, "ingredients-form-page.html", data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ingredients/%d/prices", ID), http.StatusSeeOther)
}
