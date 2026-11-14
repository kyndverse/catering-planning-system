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
	ingredientsRepo := repository.NewIngredientRepository(config.Logger)

	// setup service
	ingredientService := service.NewIngredientService(config.Logger, config.DB, config.Validator, ingredientsRepo)

	// setup handler
	renderer := web.NewRenderer()
	pageHandler := handler.NewPageHandler(renderer)
	ingredientHandler := handler.NewIngredientHandler(config.Logger, renderer, config.Decoder, ingredientService)

	// setup route
	routeConfig := &router.RouteConfig{
		App:               config.App,
		PageHandler:       pageHandler,
		IngredientHandler: ingredientHandler,
	}

	routeConfig.Setup()
}
