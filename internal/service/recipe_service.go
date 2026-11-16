package service

import (
	"context"
	"database/sql"
	"golang-web/internal/domain/entity"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/validator"

	"github.com/sirupsen/logrus"
)

type RecipesService struct {
	Log              *logrus.Logger
	DB               *sql.DB
	Validator        *validator.Validator
	RecipeRepository *repository.RecipesRepository
}

func NewRecipesService(logger *logrus.Logger, db *sql.DB, validator *validator.Validator, recipesRepository *repository.RecipesRepository) *RecipesService {
	return &RecipesService{
		Log:              logger,
		DB:               db,
		Validator:        validator,
		RecipeRepository: recipesRepository,
	}
}

func (s *RecipesService) Search(ctx context.Context, request *model.SearchRecipeRequest) ([]model.RecipeResponse, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	var recipes []entity.Recipe
	if request.Type == "" {
		recipes, err = s.RecipeRepository.FindAll(tx)
		if err != nil {
			s.Log.WithError(err).Error("failed to find recipes")
			return nil, err
		}
	} else {
		recipes, err = s.RecipeRepository.SearchByType(tx, request.Type)
		if err != nil {
			s.Log.WithError(err).Error("failed to find recipes by type")
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	response := []model.RecipeResponse{}
	for _, recipe := range recipes {
		response = append(response, model.RecipeResponse{
			ID:        recipe.ID,
			Name:      recipe.Name,
			Type:      recipe.Type,
			YieldQty:  recipe.YieldQty,
			YieldUnit: recipe.YieldUnit,
		})
	}

	return response, nil
}

func (s *RecipesService) AddRecipe(ctx context.Context, request *model.CreateRecipeRequest) (*model.RecipeResponse, validator.ValidatorErrors, error) {
	fieldErrors := s.Validator.ValidateStruct(request)
	if fieldErrors != nil {
		return nil, fieldErrors, nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, nil, err
	}
	defer tx.Rollback()

	recipe := &entity.Recipe{
		Name:      request.Name,
		Type:      request.Type,
		YieldQty:  request.YieldQty,
		YieldUnit: request.YieldUnit,
	}

	if err := s.RecipeRepository.Create(tx, recipe); err != nil {
		s.Log.WithError(err).Error("failed to create recipe")
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, err
	}

	return &model.RecipeResponse{
		ID:        recipe.ID,
		Name:      recipe.Name,
		Type:      recipe.Type,
		YieldQty:  recipe.YieldQty,
		YieldUnit: recipe.YieldUnit,
	}, nil, nil
}
