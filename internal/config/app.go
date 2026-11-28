package config

import (
	"database/sql"

	"golang-web/internal/handler"
	"golang-web/internal/middleware"
	"golang-web/internal/repository"
	"golang-web/internal/router"
	"golang-web/internal/service"
	"golang-web/internal/validator"
	"golang-web/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type BootstrapConfig struct {
	App         chi.Router
	RedisClient *redis.Client
	Logger      *logrus.Logger
	Decoder     *form.Decoder
	Validator   *validator.Validator
	DB          *sql.DB
}

func Bootstrap(config *BootstrapConfig) {
	// setup repo
	userRepo := repository.NewUserRepository(config.Logger)
	sessionRepo := repository.NewSessionRepository(config.RedisClient)
	ingredientRepo := repository.NewIngredientRepository(config.Logger)
	ingredientPriceRepo := repository.NewIngredientPriceRepository(config.Logger)
	recipeRepo := repository.NewRecipesRepository(config.Logger)
	orderRepo := repository.NewOrderRepository(config.Logger)
	productionRepo := repository.NewProductionRepository(config.Logger)
	stockRepo := repository.NewStockRepository(config.Logger)
	cogsRepo := repository.NewCogsRepository(config.Logger)
	usageReportRepo := repository.NewUsageReportRepository(config.Logger)

	// setup service
	authService := service.NewAuthService(config.Logger, config.DB, config.Validator, userRepo, sessionRepo)
	ingredientService := service.NewIngredientService(config.Logger, config.DB, config.Validator, ingredientRepo, ingredientPriceRepo)
	recipeService := service.NewRecipesService(config.Logger, config.DB, config.Validator, recipeRepo)
	orderService := service.NewOrdersService(config.Logger, config.DB, config.Validator, orderRepo)
	productionService := service.NewProductionService(config.Logger, config.DB, productionRepo, orderRepo)
	stockService := service.NewStockService(config.DB, config.Validator, stockRepo)
	cogsService := service.NewCogsService(config.Logger, config.DB, cogsRepo)
	usageReportService := service.NewUsageReportService(config.Logger, config.DB, usageReportRepo, orderRepo)

	// setup handler
	renderer := web.NewRenderer()
	pageHandler := handler.NewPageHandler(renderer, ingredientService, recipeService)
	authHandler := handler.NewAuthHandler(config.Logger, renderer, config.Decoder, authService)
	ingredientHandler := handler.NewIngredientHandler(config.Logger, renderer, config.Decoder, ingredientService)
	recipeHandler := handler.NewRecipesHandler(config.Logger, renderer, config.Decoder, recipeService, ingredientService)
	orderHandler := handler.NewOrderHandler(config.Logger, renderer, config.Decoder, orderService)
	productionHandler := handler.NewProductionHandler(config.Logger, renderer, productionService)
	stockHandler := handler.NewStockHandler(config.Logger, renderer, config.Decoder, stockService)
	cogsHandler := handler.NewCogsHandler(config.Logger, renderer, cogsService)
	usageReportHandler := handler.NewUsageReportHandler(config.Logger, renderer, usageReportService)

	// setup middleware
	authMiddleware := middleware.NewAuthMiddleware(config.Logger, sessionRepo, renderer)

	// setup route
	routeConfig := &router.RouteConfig{
		App:                config.App,
		PageHandler:        pageHandler,
		AuthHandler:        authHandler,
		IngredientHandler:  ingredientHandler,
		RecipeHandler:      recipeHandler,
		OrderHandler:       orderHandler,
		ProductionHandler:  productionHandler,
		StockHandler:       stockHandler,
		CogsHandler:        cogsHandler,
		UsageReportHandler: usageReportHandler,
		AuthMiddleware:     authMiddleware,
	}

	routeConfig.Setup()
}
