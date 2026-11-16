package repository

import (
	"database/sql"
	"golang-web/internal/domain/entity"

	"github.com/sirupsen/logrus"
)

type RecipesRepository struct {
	Log *logrus.Logger
}

func NewRecipesRepository(logger *logrus.Logger) *RecipesRepository {
	return &RecipesRepository{
		Log: logger,
	}
}

func (r *RecipesRepository) FindAll(tx *sql.Tx) ([]entity.Recipe, error) {
	query := `SELECT id, name, type, yield_qty, yield_unit FROM recipes`

	var recipes []entity.Recipe

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		recipe := entity.Recipe{}
		if err := rows.Scan(
			&recipe.ID,
			&recipe.Name,
			&recipe.Type,
			&recipe.YieldQty,
			&recipe.YieldUnit,
		); err != nil {
			return nil, err
		}

		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return recipes, nil
}

func (r *RecipesRepository) SearchByType(tx *sql.Tx, recipeType string) ([]entity.Recipe, error) {
	query := `SELECT id, name, type, yield_qty, yield_unit FROM recipes WHERE type LIKE $1`

	var recipes []entity.Recipe

	rows, err := tx.Query(query, "%"+recipeType+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		recipe := entity.Recipe{}
		if err := rows.Scan(
			&recipe.ID,
			&recipe.Name,
			&recipe.Type,
			&recipe.YieldQty,
			&recipe.YieldUnit,
		); err != nil {
			return nil, err
		}

		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return recipes, nil
}

func (p *RecipesRepository) Create(tx *sql.Tx, recipe *entity.Recipe) error {
	query := `
        INSERT INTO recipes (name, type, yield_qty, yield_unit)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	return tx.QueryRow(
		query,
		recipe.Name,
		recipe.Type,
		recipe.YieldQty,
		recipe.YieldUnit,
	).Scan(&recipe.ID)
}
