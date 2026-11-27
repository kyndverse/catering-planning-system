package service

import (
	"context"
	"database/sql"

	"golang-web/internal/entity"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/validator"

	"github.com/sirupsen/logrus"
)

type IngredientService struct {
	Log                       *logrus.Logger
	DB                        *sql.DB
	Validator                 *validator.Validator
	IngredientRepository      *repository.IngredientRepository
	IngredientPriceRepository *repository.IngredientPriceRepository
}

func NewIngredientService(logger *logrus.Logger, db *sql.DB, validator *validator.Validator, ingredientRepository *repository.IngredientRepository, ingredientPriceRepository *repository.IngredientPriceRepository) *IngredientService {
	return &IngredientService{
		Log:                       logger,
		DB:                        db,
		Validator:                 validator,
		IngredientRepository:      ingredientRepository,
		IngredientPriceRepository: ingredientPriceRepository,
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

func (p *IngredientService) GetAllPricesByIngredientID(ctx context.Context, ID int64) ([]model.IngredientPriceResponse, error) {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		p.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	prices, err := p.IngredientPriceRepository.FindAllByIngedientID(tx, ID)
	if err != nil {
		p.Log.WithError(err).Error("failed to find ingredient prices")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		p.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	response := []model.IngredientPriceResponse{}
	for _, price := range prices {
		response = append(response, model.IngredientPriceResponse{
			ID:                   price.ID,
			IngredientID:         price.IngredientID,
			PricePerPurchaseUnit: price.PricePerPurchaseUnit,
			EffectiveFrom:        price.EffectiveFrom,
		})
	}

	return response, nil
}

func (p *IngredientService) AddIngredientPrice(ctx context.Context, request *model.CreateIngredientPriceRequest) (*model.IngredientPriceResponse, validator.ValidatorErrors, error) {
	fieldErrors := p.Validator.ValidateStruct(request)
	if fieldErrors != nil {
		return nil, fieldErrors, nil
	}

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		p.Log.WithError(err).Error("failed to begin transaction")
		return nil, nil, err
	}
	defer tx.Rollback()

	price := &entity.IngredientPrice{
		IngredientID:         request.IngredientID,
		PricePerPurchaseUnit: request.PricePerPurchaseUnit,
		EffectiveFrom:        request.EffectiveFrom,
	}

	if err := p.IngredientPriceRepository.Create(tx, price); err != nil {
		p.Log.WithError(err).Error("failed to create ingredient prices")
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		p.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, err
	}

	return &model.IngredientPriceResponse{
		ID:                   price.ID,
		IngredientID:         request.IngredientID,
		PricePerPurchaseUnit: request.PricePerPurchaseUnit,
		EffectiveFrom:        request.EffectiveFrom,
	}, nil, nil
}

func (p *IngredientService) AddIngredient(ctx context.Context, request *model.CreateIngredientRequest) (*model.IngredientResponse, validator.ValidatorErrors, error) {
	fieldErrors := p.Validator.ValidateStruct(request)
	if fieldErrors != nil {
		return nil, fieldErrors, nil
	}

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		p.Log.WithError(err).Error("failed to begin transaction")
		return nil, nil, err
	}
	defer tx.Rollback()

	exists, err := p.IngredientRepository.ExistsByName(tx, request.Name)
	if err != nil {
		p.Log.WithError(err).Error("failed to check ingredient name")
		return nil, nil, err
	}

	if exists {
		formError := validator.ValidatorErrors{
			"Error": "Nama bahan sudah digunakan",
		}

		return nil, formError, nil
	}

	ingredient := &entity.Ingredient{
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

func (p *IngredientService) GetIngredientById(ctx context.Context, id int64) (*model.IngredientResponse, error) {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		p.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	ingredient := new(entity.Ingredient)
	ingredient.ID = id

	if err := p.IngredientRepository.FindOneByID(tx, ingredient); err != nil {
		p.Log.WithError(err).Error("failed to find ingredient by id")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		p.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return &model.IngredientResponse{
		ID:                     ingredient.ID,
		Name:                   ingredient.Name,
		BaseUnit:               ingredient.BaseUnit,
		PurchaseUnit:           ingredient.PurchaseUnit,
		BaseQtyPerPurchaseUnit: ingredient.BaseQtyPerPurchaseUnit,
	}, nil
}

func (p *IngredientService) EditIngredient(ctx context.Context, request *model.UpdateIngredientRequest) (*model.IngredientResponse, validator.ValidatorErrors, error) {
	fieldErrors := p.Validator.ValidateStruct(request)
	if fieldErrors != nil {
		return nil, fieldErrors, nil
	}

	tx, err := p.DB.BeginTx(ctx, nil)
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

	if err := p.IngredientRepository.Update(tx, ingredient); err != nil {
		p.Log.WithError(err).Error("failed to update ingredient")
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

func (p *IngredientService) DeleteIngredient(ctx context.Context, request *model.DeleteIngredientRequest) error {
	tx, err := p.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		p.Log.WithError(err).Error("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	ingredient := &entity.Ingredient{
		ID: request.ID,
	}

	if err := p.IngredientRepository.Delete(tx, ingredient); err != nil {
		p.Log.WithError(err).Error("failed to delete ingredient")
		return err
	}

	if err := tx.Commit(); err != nil {
		p.Log.WithError(err).Error("failed to commit transaction")
		return err
	}

	return nil
}
