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

type IngredientService struct {
	Log                  *logrus.Logger
	DB                   *sql.DB
	Validator            *validator.Validator
	IngredientRepository *repository.IngredientRepository
}

func NewIngredientService(logger *logrus.Logger, db *sql.DB, validator *validator.Validator, ingredientRepository *repository.IngredientRepository) *IngredientService {
	return &IngredientService{
		Log:                  logger,
		DB:                   db,
		Validator:            validator,
		IngredientRepository: ingredientRepository,
	}
}

func (p *IngredientService) GetAllIngredients(ctx context.Context) ([]model.IngredientResponse, error) {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		p.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	ingredients, err := p.IngredientRepository.FindAll(tx)
	if err != nil {
		p.Log.WithError(err).Error("failed to find ingredients")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		p.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	response := []model.IngredientResponse{}
	for _, ingredient := range ingredients {
		response = append(response, model.IngredientResponse{
			ID:                     ingredient.ID,
			Name:                   ingredient.Name,
			BaseUnit:               ingredient.BaseUnit,
			PurchaseUnit:           ingredient.PurchaseUnit,
			BaseQtyPerPurchaseUnit: ingredient.BaseQtyPerPurchaseUnit,
		})
	}

	return response, nil
}

func (p *IngredientService) AddIngredient(ctx context.Context, request *model.CreateIngredientRequest) (*model.IngredientResponse, validator.ValidatorErrors, error) {
	errors := p.Validator.ValidateStruct(request)
	if errors != nil {
		return nil, errors, nil
	}

	tx, err := p.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		p.Log.WithError(err).Error("failed to begin transaction")
		return nil, nil, err
	}
	defer tx.Rollback()

	ingredient := &entity.Ingredient{
		ID:                     request.ID,
		Name:                   request.Name,
		BaseUnit:               request.BaseUnit,
		PurchaseUnit:           request.PurchaseUnit,
		BaseQtyPerPurchaseUnit: request.BaseQtyPerPurchaseUnit,
	}

	if err := p.IngredientRepository.Create(tx, ingredient); err != nil {
		p.Log.WithError(err).Error("failed to create ingredient")
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		p.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, err
	}

	return &model.IngredientResponse{
		ID:                     ingredient.ID,
		Name:                   ingredient.Name,
		BaseUnit:               ingredient.BaseUnit,
		PurchaseUnit:           ingredient.PurchaseUnit,
		BaseQtyPerPurchaseUnit: ingredient.BaseQtyPerPurchaseUnit,
	}, nil, nil
}
