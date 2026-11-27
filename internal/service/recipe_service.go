package service

import (
	"context"
	"database/sql"
	"errors"

	"golang-web/internal/entity"
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

func (s *RecipesService) AddRecipeItem(ctx context.Context, request *model.AddRecipeItemRequest) (validator.ValidatorErrors, error) {
	if fieldErrors := s.Validator.ValidateStruct(request); fieldErrors != nil {
		return fieldErrors, nil
	}

	kind, targetID, ok := parseItemRef(request.Item)
	if !ok {
		return validator.ValidatorErrors{
			"Error": "Pilih bahan baku atau sub-resep yang valid",
		}, nil
	}

	quantity, ok := parsePositiveQuantity(request.Quantity)
	if !ok {
		return validator.ValidatorErrors{
			"Error": "Jumlah harus angka lebih dari 0 (maksimal 6 angka desimal)",
		}, nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	exists, err := s.RecipeRepository.RecipeExists(tx, request.RecipeID)
	if err != nil {
		s.Log.WithError(err).Error("failed to check recipe")
		return nil, err
	}
	if !exists {
		return validator.ValidatorErrors{
			"Error": "Recipe tidak ditemukan",
		}, nil
	}

	message, err := s.checkItemRules(tx, request.RecipeID, kind, targetID)
	if err != nil {
		return nil, err
	}
	if message != "" {
		return validator.ValidatorErrors{
			"Error": message,
		}, nil
	}

	if err := s.RecipeRepository.AddItem(tx, request.RecipeID, kind, targetID, quantity); err != nil {
		s.Log.WithError(err).Error("failed to add recipe item")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return nil, nil
}

func (s *RecipesService) UpdateRecipeItem(ctx context.Context, request *model.UpdateRecipeItemRequest) (validator.ValidatorErrors, error) {
	if fieldErrors := s.Validator.ValidateStruct(request); fieldErrors != nil {
		return fieldErrors, nil
	}

	quantity, ok := parsePositiveQuantity(request.Quantity)
	if !ok {
		return validator.ValidatorErrors{
			"Error": "Jumlah harus angka lebih dari 0 (maksimal 6 angka desimal)",
		}, nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	updated, err := s.RecipeRepository.UpdateItemQuantity(tx, request.RecipeID, request.ItemID, quantity)
	if err != nil {
		s.Log.WithError(err).Error("failed to update recipe item")
		return nil, err
	}
	if !updated {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return nil, nil
}

func (s *RecipesService) DeleteRecipeItem(ctx context.Context, request *model.DeleteRecipeItemRequest) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	deleted, err := s.RecipeRepository.DeleteItem(tx, request.RecipeID, request.ItemID)
	if err != nil {
		s.Log.WithError(err).Error("failed to delete recipe item")
		return err
	}
	if !deleted {
		return err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return err
	}

	return nil
}

// Rules Mapping
// checkItemRules memeriksa aturan bisnis. Mengembalikan pesan untuk pengguna
func (s *RecipesService) checkItemRules(tx *sql.Tx, recipeID int64, kind string, targetID int64) (string, error) {
	switch kind {
	case model.ItemKindIngredient:
		exists, err := s.RecipeRepository.IngredientExists(tx, targetID)
		if err != nil {
			s.Log.WithError(err).Error("failed to check ingredient")
			return "", err
		}
		if !exists {
			return "Bahan baku tidak ditemukan", nil
		}

	case model.ItemKindRecipe:
		if targetID == recipeID {
			return "Resep tidak boleh memuat dirinya sendiri", nil
		}

		subType, err := s.RecipeRepository.FindType(tx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "Sub-resep tidak ditemukan", nil
		}
		if err != nil {
			s.Log.WithError(err).Error("failed to find sub recipe")
			return "", err
		}
		if subType == "MENU" {
			return "Menu tidak boleh dijadikan sub-resep", nil
		}

		cycle, err := s.RecipeRepository.WouldCreateCycle(tx, recipeID, targetID)
		if err != nil {
			s.Log.WithError(err).Error("failed to check recipe cycle")
			return "", err
		}
		if cycle {
			return "Sub-resep ini akan membuat siklus (resep saling memuat)", nil
		}
	}

	// Satu bahan/sub-resep hanya boleh muncul sekali per resep; ubah jumlahnya
	// di tabel Isi Resep. Hapus blok ini jika Anda ingin mengizinkan duplikat.
	dup, err := s.RecipeRepository.ItemExists(tx, recipeID, kind, targetID)
	if err != nil {
		s.Log.WithError(err).Error("failed to check duplicate item")
		return "", err
	}
	if dup {
		return "Isi ini sudah ada di resep. Ubah jumlahnya di tabel Isi Resep.", nil
	}

	return "", nil
}
