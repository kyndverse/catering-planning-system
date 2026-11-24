package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"golang-web/internal/handler"
	"golang-web/web"
)

func TestOrdersNewRoutePrecedesOrderID(t *testing.T) {
	mux := chi.NewRouter()
	orders := handler.NewOrderHandler(nil, web.NewRenderer(), nil, nil)
	(&RouteConfig{App: mux, OrderHandler: orders}).Setup()
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/orders/new", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Tambah Pesanan") {
		t.Fatalf("GET /orders/new: status %d, body %q", response.Code, response.Body.String())
	}
}
