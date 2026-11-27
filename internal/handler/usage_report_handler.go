package handler

import (
	"context"
	"golang-web/internal/service"
	"golang-web/internal/timeutil"
	"golang-web/internal/validator"
	"golang-web/internal/view"
	"golang-web/web"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type UsageReportHandler struct {
	Log                *logrus.Logger
	Renderer           *web.Renderer
	UsageReportService *service.UsageReportService
}

func NewUsageReportHandler(log *logrus.Logger, renderer *web.Renderer, s *service.UsageReportService) *UsageReportHandler {
	return &UsageReportHandler{
		Log:                log,
		Renderer:           renderer,
		UsageReportService: s,
	}
}

func (h *UsageReportHandler) Index(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	raw := r.URL.Query().Get("date")

	if raw != "" {
		date, err := timeutil.ParseDateISO(raw)
		if err == nil {
			http.Redirect(w, r, "/reports/usage/"+timeutil.FormatDateISO(date), http.StatusSeeOther)
			return
		}

		h.renderIndex(w, r.Context(), now, validator.ValidatorErrors{"Date": "Format tanggal tidak valid"})
		return
	}

	h.renderIndex(w, r.Context(), now, nil)
}

func (h *UsageReportHandler) renderIndex(w http.ResponseWriter, ctx context.Context, now time.Time, errs validator.ValidatorErrors) {
	rows, err := h.UsageReportService.ListUsageDates(ctx, now)
	if err != nil {
		h.fail(w, err)
		return
	}

	h.Renderer.Render(w, "reports/usage_index.html", map[string]any{
		"title":  "Laporan Pemakaian",
		"active": "reports",
		"errors": errs,
		"date":   timeutil.FormatDateISO(timeutil.Today(now)),
		"dates":  view.UsageDates(rows),
	})
}

func (h *UsageReportHandler) Show(w http.ResponseWriter, r *http.Request) {
	date, err := timeutil.ParseDateISO(r.PathValue("date"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	report, err := h.UsageReportService.GetUsageReport(r.Context(), date, time.Now())
	if err != nil {
		h.fail(w, err)
		return
	}

	production, rows, summary := view.UsageReport(report)
	h.Renderer.Render(w, "reports/usage.html", map[string]any{
		"title":      "Laporan Pemakaian",
		"active":     "reports",
		"errors":     nil,
		"production": production,
		"rows":       rows,
		"summary":    summary})
}

func (h *UsageReportHandler) fail(w http.ResponseWriter, err error) {
	h.Log.WithError(err).Error("usage report request failed")
	http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
}
