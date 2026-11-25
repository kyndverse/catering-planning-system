package handler

import (
	"context"
	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/service"
	"golang-web/internal/timeutil"
	"golang-web/internal/view"
	"golang-web/web"
	"net/http"
	"strconv"
	"time"
)

type StockHandler struct {
	Log          *logrus.Logger
	Renderer     *web.Renderer
	Decoder      *form.Decoder
	StockService *service.StockService
}

func NewStockHandler(l *logrus.Logger, r *web.Renderer, d *form.Decoder, s *service.StockService) *StockHandler {
	return &StockHandler{l, r, d, s}
}
func (h *StockHandler) Index(w http.ResponseWriter, r *http.Request) {
	rows, n, e := h.StockService.ListStocks(r.Context(), model.SearchStockRequest{Q: r.URL.Query().Get("q")})
	if e != nil {
		h.serverError(w, e)
		return
	}
	h.Renderer.Render(w, "stocks/index.html", map[string]any{"title": "Stok", "active": "stock", "errors": nil, "negativeCount": n, "filter": model.SearchStockRequest{Q: r.URL.Query().Get("q")}, "stocks": view.StockRows(rows)})
}
func (h *StockHandler) Movements(w http.ResponseWriter, r *http.Request) {
	req := model.SearchMovementRequest{IngredientID: r.URL.Query().Get("ingredient_id"), Type: r.URL.Query().Get("type"), From: r.URL.Query().Get("from"), To: r.URL.Query().Get("to"), Page: r.URL.Query().Get("page")}
	rows, p, pages, fe, e := h.StockService.ListMovements(r.Context(), req, time.Now())
	if e != nil {
		h.serverError(w, e)
		return
	}
	f := normalizedMovementFilter(req)
	opts, e := h.StockService.ListIngredientOptions(r.Context())
	if e != nil {
		h.serverError(w, e)
		return
	}
	filter := model.SearchMovementRequest{IngredientID: f.IngredientID, Type: f.Type, From: f.From, To: f.To}
	h.Renderer.Render(w, "stocks/movements.html", map[string]any{"title": "Riwayat Stok", "active": "stock", "errors": fe, "hasFilter": f.IngredientID != "" || f.Type != "" || f.From != "" || f.To != "", "filter": filter, "ingredientOptions": view.IngredientOptionRows(opts), "movements": view.MovementRows(rows), "pagination": view.Pagination{Page: p, TotalPages: pages, HasPrev: p > 1, HasNext: p < pages, PrevURL: view.BuildMovementsURL(f, p-1), NextURL: view.BuildMovementsURL(f, p+1)}})
}

// normalizedMovementFilter mirrors the service's permissive filter handling so
// pagination preserves valid filters when another query parameter is invalid.
func normalizedMovementFilter(req model.SearchMovementRequest) repository.MovementFilter {
	f := repository.MovementFilter{}
	if id, err := strconv.ParseInt(req.IngredientID, 10, 64); err == nil && id > 0 {
		f.IngredientID = req.IngredientID
	}
	if req.Type == "PURCHASE" || req.Type == "USAGE" || req.Type == "ADJUSTMENT" {
		f.Type = req.Type
	}
	if _, err := timeutil.ParseDateISO(req.From); err == nil && req.From != "" {
		f.From = req.From
	}
	if _, err := timeutil.ParseDateISO(req.To); err == nil && req.To != "" {
		f.To = req.To
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		f.From, f.To = "", ""
	}
	return f
}
func (h *StockHandler) NewPurchasePage(w http.ResponseWriter, r *http.Request) {
	h.renderPurchase(w, model.CreatePurchaseRequest{MovementDate: timeutil.FormatDateISO(timeutil.Today(time.Now())), IngredientID: r.URL.Query().Get("ingredient_id")}, nil)
}
func (h *StockHandler) NewUsagePage(w http.ResponseWriter, r *http.Request) {
	today := timeutil.FormatDateISO(timeutil.Today(time.Now()))
	h.renderUsage(w, model.CreateUsageRequest{MovementDate: today, ProductionDate: today, IngredientID: r.URL.Query().Get("ingredient_id")}, nil, "")
}
func (h *StockHandler) NewAdjustmentPage(w http.ResponseWriter, r *http.Request) {
	h.renderAdjustment(w, model.CreateAdjustmentRequest{MovementDate: timeutil.FormatDateISO(timeutil.Today(time.Now())), IngredientID: r.URL.Query().Get("ingredient_id")}, nil, "")
}
func (h *StockHandler) CreatePurchase(w http.ResponseWriter, r *http.Request) {
	var x model.CreatePurchaseRequest
	if !h.decode(w, r, &x) {
		return
	}
	res, fe, e := h.StockService.RecordPurchase(r.Context(), x, time.Now())
	if e != nil {
		h.serverError(w, e)
		return
	}
	if fe != nil {
		h.renderPurchase(w, x, fe)
		return
	}
	http.Redirect(w, r, "/stock/movements?ingredient_id="+strconv.FormatInt(res.IngredientID, 10), http.StatusSeeOther)
}
func (h *StockHandler) CreateUsage(w http.ResponseWriter, r *http.Request) {
	var x model.CreateUsageRequest
	if !h.decode(w, r, &x) {
		return
	}
	res, fe, e := h.StockService.RecordUsage(r.Context(), x, time.Now())
	if e != nil {
		h.serverError(w, e)
		return
	}
	if fe != nil {
		h.renderUsage(w, x, fe, "")
		return
	}
	if res.Warning != "" {
		x.ConfirmNegative = "1"
		h.renderUsage(w, x, nil, res.Warning)
		return
	}
	http.Redirect(w, r, "/stock/movements?ingredient_id="+strconv.FormatInt(res.IngredientID, 10), http.StatusSeeOther)
}
func (h *StockHandler) CreateAdjustment(w http.ResponseWriter, r *http.Request) {
	var x model.CreateAdjustmentRequest
	if !h.decode(w, r, &x) {
		return
	}
	res, fe, e := h.StockService.RecordAdjustment(r.Context(), x, time.Now())
	if e != nil {
		h.serverError(w, e)
		return
	}
	if fe != nil {
		h.renderAdjustment(w, x, fe, "")
		return
	}
	if res.Warning != "" {
		x.ConfirmNegative = "1"
		h.renderAdjustment(w, x, nil, res.Warning)
		return
	}
	http.Redirect(w, r, "/stock/movements?ingredient_id="+strconv.FormatInt(res.IngredientID, 10), http.StatusSeeOther)
}
func (h *StockHandler) decode(w http.ResponseWriter, r *http.Request, x any) bool {
	if e := r.ParseForm(); e != nil {
		http.Error(w, "invalid request body", 400)
		return false
	}
	if e := h.Decoder.Decode(x, r.PostForm); e != nil {
		http.Error(w, "Invalid form data", 400)
		return false
	}
	return true
}
func (h *StockHandler) renderPurchase(w http.ResponseWriter, x model.CreatePurchaseRequest, e map[string]string) {
	opts, err := h.StockService.ListIngredientOptions(context.Background())
	if err != nil {
		h.serverError(w, err)
		return
	}
	h.Renderer.Render(w, "stocks/purchase.html", map[string]any{"title": "Catat Pembelian", "active": "stock", "today": timeutil.FormatDateISO(timeutil.Today(time.Now())), "ingredientOptions": view.IngredientOptionRows(opts), "form": model.FormData[model.CreatePurchaseRequest]{Action: "/stock/purchases", Data: x, Errors: e}, "submitLabel": "Catat Pembelian", "cancelURL": "/stock"})
}
func (h *StockHandler) renderUsage(w http.ResponseWriter, x model.CreateUsageRequest, e map[string]string, warn string) {
	opts, err := h.StockService.ListIngredientOptions(context.Background())
	if err != nil {
		h.serverError(w, err)
		return
	}
	label := "Catat Pemakaian"
	if warn != "" {
		label = "Tetap Catat"
	}
	h.Renderer.Render(w, "stocks/usage.html", map[string]any{"title": "Catat Pemakaian", "active": "stock", "today": timeutil.FormatDateISO(timeutil.Today(time.Now())), "ingredientOptions": view.IngredientOptionRows(opts), "form": model.FormData[model.CreateUsageRequest]{Action: "/stock/usages", Data: x, Errors: e}, "warning": warn, "submitLabel": label, "cancelURL": "/stock"})
}
func (h *StockHandler) renderAdjustment(w http.ResponseWriter, x model.CreateAdjustmentRequest, e map[string]string, warn string) {
	opts, err := h.StockService.ListIngredientOptions(context.Background())
	if err != nil {
		h.serverError(w, err)
		return
	}
	label := "Catat Koreksi"
	if warn != "" {
		label = "Tetap Catat"
	}
	h.Renderer.Render(w, "stocks/adjustment.html", map[string]any{"title": "Koreksi Stok", "active": "stock", "today": timeutil.FormatDateISO(timeutil.Today(time.Now())), "ingredientOptions": view.IngredientOptionRows(opts), "form": model.FormData[model.CreateAdjustmentRequest]{Action: "/stock/adjustments", Data: x, Errors: e}, "warning": warn, "submitLabel": label, "cancelURL": "/stock"})
}
func (h *StockHandler) serverError(w http.ResponseWriter, e error) {
	h.Log.WithError(e).Error("stock request failed")
	http.Error(w, "Internal server error", 500)
}
