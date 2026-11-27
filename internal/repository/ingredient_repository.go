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
	query := `SELECT id, name, base_unit, purchase_unit, base_qty_per_purchase_unit FROM ingredients`

	var ingredients []entity.Ingredient

	rows, err := tx.Query(query)
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

func (p *IngredientRepository) FindOneByID(tx *sql.Tx, ingredient *entity.Ingredient) error {
	query := `SELECT id, name, base_unit, purchase_unit, base_qty_per_purchase_unit FROM ingredients WHERE id=$1 LIMIT 1`

	return tx.QueryRow(
		query,
		ingredient.ID,
	).Scan(
		&ingredient.ID,
		&ingredient.Name,
		&ingredient.BaseUnit,
		&ingredient.PurchaseUnit,
		&ingredient.BaseQtyPerPurchaseUnit)
}

func (p *IngredientRepository) ExistsByName(tx *sql.Tx, name string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM ingredients WHERE name=$1)`

	var exists bool
	if err := tx.QueryRow(query, name).Scan(&exists); err != nil {
		return false, nil
	}

	return exists, nil
}

func (p *IngredientRepository) Update(tx *sql.Tx, ingredient *entity.Ingredient) error {
	query := `UPDATE ingredients SET name=$1, base_unit=$2, purchase_unit=$3, base_qty_per_purchase_unit=$4 WHERE id=$5 RETURNING id, name, base_unit, purchase_unit, base_qty_per_purchase_unit`

	return tx.QueryRow(
		query,
		ingredient.Name,
		ingredient.BaseUnit,
		ingredient.PurchaseUnit,
		ingredient.BaseQtyPerPurchaseUnit,
		ingredient.ID,
	).Scan(
		&ingredient.ID,
		&ingredient.Name,
		&ingredient.BaseUnit,
		&ingredient.PurchaseUnit,
		&ingredient.BaseQtyPerPurchaseUnit)
}

func (p *IngredientRepository) Delete(tx *sql.Tx, ingredient *entity.Ingredient) error {
	query := `DELETE FROM ingredients WHERE id=$1`

	_, err := tx.Exec(query, ingredient.ID)
	if err != nil {
		return err
	}

	return nil
}
