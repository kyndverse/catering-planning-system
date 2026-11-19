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
	RecipeRepository *repository.RecipeRepository
}

func NewRecipesService(logger *logrus.Logger, db *sql.DB, validator *validator.Validator, recipesRepository *repository.RecipeRepository) *RecipesService {
	return &RecipesService{
		Log:              logger,
		DB:               db,
		Validator:        validator,
		RecipeRepository: recipesRepository,
	}
}

func (s *RecipesService) Search(ctx context.Context, request *model.SearchRecipeRequest) ([]model.RecipeSummary, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	recipeSummries, err := s.RecipeRepository.ListSummaries(tx, request.Type)
	if err != nil {
		s.Log.WithError(err).Error("failed to find recipes")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return recipeSummries, nil
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

func (s *RecipesService) EditRecipe(ctx context.Context, request *model.UpdateRecipeRequest) (*model.RecipeResponse, validator.ValidatorErrors, error) {
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
		ID:        request.ID,
		Name:      request.Name,
		Type:      request.Type,
		YieldQty:  request.YieldQty,
		YieldUnit: request.YieldUnit,
	}

	if err := s.RecipeRepository.Update(tx, recipe); err != nil {
		s.Log.WithError(err).Error("failed to update recipe")
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
		YieldQty:  request.YieldQty,
		YieldUnit: recipe.YieldUnit,
	}, nil, nil
}

func (s *RecipesService) DeleteRecipe(ctx context.Context, request *model.DeleteRecipeRequest) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	recipe := &entity.Recipe{
		ID: request.ID,
	}

	if err := s.RecipeRepository.Delete(tx, recipe); err != nil {
		s.Log.WithError(err).Error("failed to delete recipe")
		return err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return err
	}

	return nil
}

func (s *RecipesService) GetRecipeById(ctx context.Context, ID int64) (*model.RecipeResponse, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	recipe := &entity.Recipe{
		ID: ID,
	}

	if err := s.RecipeRepository.FindOne(tx, recipe); err != nil {
		s.Log.WithError(err).Error("failed to get recipe")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	response := &model.RecipeResponse{
		ID:        recipe.ID,
		Name:      recipe.Name,
		Type:      recipe.Type,
		YieldQty:  recipe.YieldQty,
		YieldUnit: recipe.YieldUnit,
	}

	return response, nil
}

func (s *RecipesService) ListItemsByRecipeID(ctx context.Context, ID int64) ([]model.RecipeItemView, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	items, err := s.RecipeRepository.ListItemsByRecipeID(tx, ID)
	if err != nil {
		s.Log.WithError(err).Error("failed to get recipe items")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return items, nil
}

func (s *RecipesService) ListSubRecipeOptions(ctx context.Context, ID int64) ([]model.SubRecipeItemOption, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	items, err := s.RecipeRepository.ListSubRecipeOptions(tx, ID)
	if err != nil {
		s.Log.WithError(err).Error("failed to get sub recipe option")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return items, nil
}

func (s *RecipesService) ExplodeToIngredients(ctx context.Context, ID int64, qty string) ([]model.IngredientRequirement, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	reqs, err := s.RecipeRepository.ExplodeToIngredients(tx, ID, qty)
	if err != nil {
		s.Log.WithError(err).Error("failed to explode recipe")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return reqs, nil
}
