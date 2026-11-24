package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"golang-web/internal/service"
	"golang-web/internal/timeutil"
	"golang-web/internal/validator"
	"golang-web/internal/view"
	"golang-web/web"
)

type ProductionHandler struct {
	Log               *logrus.Logger
	Renderer          *web.Renderer
	ProductionService *service.ProductionService
}

func NewProductionHandler(log *logrus.Logger, renderer *web.Renderer, production *service.ProductionService) *ProductionHandler {
	return &ProductionHandler{Log: log, Renderer: renderer, ProductionService: production}
}
func (h *ProductionHandler) Index(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	raw := r.URL.Query().Get("date")
	if raw != "" {
		date, err := timeutil.ParseDateISO(raw)
		if err == nil {
			http.Redirect(w, r, "/production/"+timeutil.FormatDateISO(date), http.StatusSeeOther)
			return
		}
		h.renderIndex(w, r.Context(), now, validator.ValidatorErrors{"Date": "Format tanggal tidak valid"})
		return
	}
	h.renderIndex(w, r.Context(), now, nil)
}
func (h *ProductionHandler) renderIndex(w http.ResponseWriter, ctx context.Context, now time.Time, errs validator.ValidatorErrors) {
	rows, err := h.ProductionService.ListUpcoming(ctx, now)
	if err != nil {
		h.fail(w, err, "gagal mengambil daftar tanggal produksi")
		return
	}
	h.Renderer.Render(w, "production/index.html", map[string]any{"title": "Produksi & Belanja", "active": "production", "errors": errs, "date": timeutil.FormatDateISO(timeutil.Today(now)), "dates": view.ProductionDates(rows)})
}

func (h *ProductionHandler) Show(w http.ResponseWriter, r *http.Request) {
	date, err := timeutil.ParseDateISO(r.PathValue("date"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	plan, err := h.ProductionService.GetPlan(r.Context(), date, time.Now())
	if err != nil {
		h.fail(w, err, "gagal mengambil rencana produksi")
		return
	}
	production, orders, menus, requirements, shopping := view.ProductionPlan(plan)
	h.Renderer.Render(w, "production/show.html", map[string]any{"title": "Produksi & Belanja", "active": "production", "errors": nil, "production": production, "orders": orders, "menus": menus, "requirements": requirements, "shopping": shopping})
}
func (h *ProductionHandler) fail(w http.ResponseWriter, err error, message string) {
	h.Log.WithError(err).Error(message)
	http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
}
