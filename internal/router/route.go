package router

import (
	"net/http"

	"golang-web/internal/handler"

	"github.com/go-chi/chi/v5"
)

type RouteConfig struct {
	App               chi.Router
	PageHandler       *handler.PageHandler
	IngredientHandler *handler.IngredientHandler
}

func (c *RouteConfig) Setup() {
	fs := http.FileServer(http.Dir("web/static"))
	c.App.Handle("/web/static/*", http.StripPrefix("/web/static/", fs))

	// Pages
	c.App.Get("/", c.PageHandler.HomePage)

	// Ingredients
	c.App.Route("/ingredients", func(r chi.Router) {
		r.Get("/", c.IngredientHandler.GetAllIngredient)
		r.Post("/", c.IngredientHandler.CreateIngredient)
		r.Get("/create", c.PageHandler.CreateIngredientForm)
	})
}
