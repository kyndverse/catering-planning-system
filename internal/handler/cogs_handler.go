package handler

import (
	"errors"
	"github.com/sirupsen/logrus"
	"golang-web/internal/service"
	"golang-web/internal/timeutil"
	"golang-web/internal/view"
	"golang-web/web"
	"net/http"
	"strconv"
	"time"
)

type CogsHandler struct {
	Log         *logrus.Logger
	Renderer    *web.Renderer
	CogsService *service.CogsService
}

func NewCogsHandler(log *logrus.Logger, renderer *web.Renderer, s *service.CogsService) *CogsHandler {
	return &CogsHandler{log, renderer, s}
}
func (h *CogsHandler) Index(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	items, err := h.CogsService.ListMenuCogs(r.Context(), now)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.Renderer.Render(w, "cogs/index.html", map[string]any{"title": "HPP", "active": "cogs", "errors": nil, "asOf": timeutil.FormatDateID(timeutil.Today(now)), "menus": view.CogsMenus(items)})
}
func (h *CogsHandler) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("menuID"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	now := time.Now()
	item, err := h.CogsService.GetMenuCogs(r.Context(), id, now)
	if errors.Is(err, service.ErrMenuNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	menu, rows := view.CogsDetail(item)
	h.Renderer.Render(w, "cogs/show.html", map[string]any{"title": "HPP", "active": "cogs", "errors": nil, "asOf": timeutil.FormatDateID(timeutil.Today(now)), "menu": menu, "rows": rows})
}
func (h *CogsHandler) fail(w http.ResponseWriter, err error) {
	h.Log.WithError(err).Error("cogs request failed")
	http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
}
