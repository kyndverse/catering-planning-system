package router

import (
	"net/http"

	"golang-web/internal/auth"
	"golang-web/internal/handler"
	"golang-web/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type RouteConfig struct {
	App                chi.Router
	PageHandler        *handler.PageHandler
	AuthHandler        *handler.AuthHandler
	IngredientHandler  *handler.IngredientHandler
	RecipeHandler      *handler.RecipeHandler
	OrderHandler       *handler.OrderHandler
	ProductionHandler  *handler.ProductionHandler
	StockHandler       *handler.StockHandler
	CogsHandler        *handler.CogsHandler
	UsageReportHandler *handler.UsageReportHandler
	AuthMiddleware     *middleware.AuthMiddleware
}

func (c *RouteConfig) Setup() {
	fs := http.FileServer(http.Dir("web/static"))
	c.App.Handle("/web/static/*", http.StripPrefix("/web/static/", fs))

	c.SetupAuthRoutes()
	c.SetupGuestRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Get("/", c.PageHandler.HomePage)

	c.App.Get("/auth/login", c.AuthHandler.LoginForm)
	c.App.Post("/auth/login", c.AuthHandler.Login)
	c.App.Post("/auth/logout", c.AuthHandler.Logout)
}

func (c *RouteConfig) SetupAuthRoutes() {
	c.App.Group(func(r chi.Router) {
		r.Use(c.AuthMiddleware.Authenticate)
		
		// Ingredient
		r.Route("/ingredients", func(r chi.Router) {
			r.Get("/", c.IngredientHandler.GetAllIngredient)
			r.Post("/", c.IngredientHandler.CreateIngredient)
			r.Get("/create", c.PageHandler.CreateIngredientForm)
			r.Get("/{id}/edit", c.PageHandler.EditIngredientForm)
			r.Post("/{id}/edit", c.IngredientHandler.EditIngredient)
			r.Post("/{id}/delete", c.IngredientHandler.DeleteIngredient)
			r.Get("/{id}/prices", c.PageHandler.IngredientPricePage)
			r.Post("/{id}/prices", c.IngredientHandler.AddIngredientPrice)
		})

		// Recipe
		r.Route("/recipes", func(r chi.Router) {
			r.Get("/", c.RecipeHandler.Search)
			r.Post("/", c.RecipeHandler.CreateRecipe)
			r.Get("/create", c.PageHandler.CreateRecipeForm)
			r.Post("/{id}/delete", c.RecipeHandler.DeleteRecipe)
			r.Get("/{id}/edit", c.PageHandler.EditRecipeForm)
			r.Post("/{id}/edit", c.RecipeHandler.EditRecipe)
			r.Get("/{id}", c.RecipeHandler.GetRecipeDetail)
			r.Post("/{id}/items", c.RecipeHandler.AddRecipeItem)
			r.Post("/{id}/items/{itemID}", c.RecipeHandler.UpdateRecipeItem)
			r.Post("/{id}/items/{itemID}/delete", c.RecipeHandler.DeleteRecipeItem)
		})

		// Order
		r.Route("/orders", func(r chi.Router) {
			r.Get("/", c.OrderHandler.Search)
			r.Get("/new", c.OrderHandler.NewOrderPage)
			r.Post("/", c.OrderHandler.CreateOrder)
			r.Get("/{id}/edit", c.OrderHandler.EditOrderPage)
			r.Post("/{id}/delete", c.OrderHandler.DeleteOrder)
			r.Post("/{id}/items", c.OrderHandler.AddOrderItem)
			r.With(c.AuthMiddleware.Authorize(auth.RoleOwner)).Post("/{id}/items/{menuID}/delete", c.OrderHandler.DeleteOrderItem)
			r.Post("/{id}/items/{menuID}", c.OrderHandler.UpdateOrderItem)
			r.Get("/{id}", c.OrderHandler.GetOrderDetail)
			r.Post("/{id}", c.OrderHandler.UpdateOrder)
		})

		// HPP production and report
		r.Get("/production", c.ProductionHandler.Index)
		r.Get("/production/{date}", c.ProductionHandler.Show)
		r.Group(func(r chi.Router) {
			r.Use(c.AuthMiddleware.Authorize(auth.RoleOwner))

			r.Get("/cogs", c.CogsHandler.Index)
			r.Get("/cogs/{menuID}", c.CogsHandler.Show)
			r.Get("/reports/usage", c.UsageReportHandler.Index)
			r.Get("/reports/usage/{date}", c.UsageReportHandler.Show)
		})

		// Stock
		r.Route("/stock", func(r chi.Router) {
			r.Get("/", c.StockHandler.Index)
			r.Get("/movements", c.StockHandler.Movements)
			r.Get("/purchases/new", c.StockHandler.NewPurchasePage)
			r.Post("/purchases", c.StockHandler.CreatePurchase)
			r.Get("/usages/new", c.StockHandler.NewUsagePage)
			r.Post("/usages", c.StockHandler.CreateUsage)
			r.Get("/adjustments/new", c.StockHandler.NewAdjustmentPage)
			r.Post("/adjustments", c.StockHandler.CreateAdjustment)
		})
	})
}
