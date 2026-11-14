package repository

import (
	"database/sql"
	"golang-web/internal/domain/entity"

	"github.com/sirupsen/logrus"
)

type IngredientRepository struct {
	Log *logrus.Logger
}

func NewIngredientRepository(logger *logrus.Logger) *IngredientRepository {
	return &IngredientRepository{
		Log: logger,
	}
}

func (p *IngredientRepository) FindAll(tx *sql.Tx) ([]entity.Ingredient, error) {
	var ingredients []entity.Ingredient

	rows, err := tx.Query("SELECT id, name, base_unit, purchase_unit, base_qty_per_purchase_unit FROM ingredients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ingredient := entity.Ingredient{}
		if err := rows.Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.BaseUnit,
			&ingredient.PurchaseUnit,
			&ingredient.BaseQtyPerPurchaseUnit,
		); err != nil {
			return nil, err
		}

		ingredients = append(ingredients, ingredient)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ingredients, nil
}

func (p *IngredientRepository) Create(tx *sql.Tx, ingredient *entity.Ingredient) error {
	query := `
        INSERT INTO ingredients (name, base_unit, purchase_unit, base_qty_per_purchase_unit)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	return tx.QueryRow(
		query,
		ingredient.Name,
		ingredient.BaseUnit,
		ingredient.PurchaseUnit,
		ingredient.BaseQtyPerPurchaseUnit,
	).Scan(&ingredient.ID)
}
