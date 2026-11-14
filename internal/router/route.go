package router

import (
	"net/http"

	"golang-web/internal/handler"

	"github.com/go-chi/chi/v5"
)

type RouteConfig struct {
	App            chi.Router
	PageHandler    *handler.PageHandler
	ProductHandler *handler.ProductHandler
}

func (c *RouteConfig) Setup() {
	fs := http.FileServer(http.Dir("web/static"))
	c.App.Handle("/web/static/*", http.StripPrefix("/web/static/", fs))

	// Pages
	c.App.Get("/", c.PageHandler.HomePage)

	// Products
	c.App.Route("/products", func(r chi.Router) {
		r.Get("/", c.ProductHandler.GetProductsHandler)
		r.Post("/", c.ProductHandler.CreateProductHandler)
		r.Get("/create", c.PageHandler.CreateProductForm)
	})
}
