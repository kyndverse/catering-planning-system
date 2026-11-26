package router

import (
	"net/http"

	"golang-web/internal/handler"

	"github.com/go-chi/chi/v5"
)

type RouteConfig struct {
	App                chi.Router
	PageHandler        *handler.PageHandler
	IngredientHandler  *handler.IngredientHandler
	RecipeHandler      *handler.RecipeHandler
	OrderHandler       *handler.OrderHandler
	ProductionHandler  *handler.ProductionHandler
	StockHandler       *handler.StockHandler
	CogsHandler        *handler.CogsHandler
	UsageReportHandler *handler.UsageReportHandler
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
		r.Get("/{id}/edit", c.PageHandler.EditIngredientForm)
		r.Post("/{id}/edit", c.IngredientHandler.EditIngredient)
		r.Post("/{id}/delete", c.IngredientHandler.DeleteIngredient)
		r.Get("/{id}/prices", c.PageHandler.IngredientPricePage)
		r.Post("/{id}/prices", c.IngredientHandler.AddIngredientPrice)
	})

	// Ingredients
	c.App.Route("/recipes", func(r chi.Router) {
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

	c.App.Route("/orders", func(r chi.Router) {
		r.Get("/", c.OrderHandler.Search)
		r.Get("/new", c.OrderHandler.NewOrderPage)
		r.Post("/", c.OrderHandler.CreateOrder)
		r.Get("/{id}/edit", c.OrderHandler.EditOrderPage)
		r.Post("/{id}/delete", c.OrderHandler.DeleteOrder)
		r.Post("/{id}/items", c.OrderHandler.AddOrderItem)
		r.Post("/{id}/items/{menuID}/delete", c.OrderHandler.DeleteOrderItem)
		r.Post("/{id}/items/{menuID}", c.OrderHandler.UpdateOrderItem)
		r.Get("/{id}", c.OrderHandler.GetOrderDetail)
		r.Post("/{id}", c.OrderHandler.UpdateOrder)
	})

	c.App.Get("/production", c.ProductionHandler.Index)
	c.App.Get("/production/{date}", c.ProductionHandler.Show)
	c.App.Get("/cogs", c.CogsHandler.Index)
	c.App.Get("/cogs/{menuID}", c.CogsHandler.Show)
	c.App.Get("/reports/usage", c.UsageReportHandler.Index)
	c.App.Get("/reports/usage/{date}", c.UsageReportHandler.Show)

	c.App.Get("/stock", c.StockHandler.Index)
	c.App.Get("/stock/movements", c.StockHandler.Movements)
	c.App.Get("/stock/purchases/new", c.StockHandler.NewPurchasePage)
	c.App.Post("/stock/purchases", c.StockHandler.CreatePurchase)
	c.App.Get("/stock/usages/new", c.StockHandler.NewUsagePage)
	c.App.Post("/stock/usages", c.StockHandler.CreateUsage)
	c.App.Get("/stock/adjustments/new", c.StockHandler.NewAdjustmentPage)
	c.App.Post("/stock/adjustments", c.StockHandler.CreateAdjustment)
}
