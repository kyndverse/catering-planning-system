package config

import (
	"database/sql"

	"golang-web/internal/handler"
	"golang-web/internal/repository"
	"golang-web/internal/router"
	"golang-web/internal/service"
	"golang-web/internal/validator"
	"golang-web/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
)

type BootstrapConfig struct {
	App       chi.Router
	Logger    *logrus.Logger
	Decoder   *form.Decoder
	Validator *validator.Validator
	DB        *sql.DB
}

func Bootstrap(config *BootstrapConfig) {
	// setup repo
	ingredientRepo := repository.NewIngredientRepository(config.Logger)
	ingredientPriceRepo := repository.NewIngredientPriceRepository(config.Logger)
	recipeRepo := repository.NewRecipesRepository(config.Logger)

	// setup service
	ingredientService := service.NewIngredientService(config.Logger, config.DB, config.Validator, ingredientRepo, ingredientPriceRepo)
	recipeService := service.NewRecipesService(config.Logger, config.DB, config.Validator, recipeRepo)

	// setup handler
	renderer := web.NewRenderer()
	pageHandler := handler.NewPageHandler(renderer, ingredientService, recipeService)
	ingredientHandler := handler.NewIngredientHandler(config.Logger, renderer, config.Decoder, ingredientService)
	recipeHandler := handler.NewRecipesHandler(config.Logger, renderer, config.Decoder, recipeService)

	// setup route
	routeConfig := &router.RouteConfig{
		App:               config.App,
		PageHandler:       pageHandler,
		IngredientHandler: ingredientHandler,
		RecipeHandler:     recipeHandler,
	}

	routeConfig.Setup()
}
