package repository

import (
	"database/sql"
	"golang-web/internal/domain/entity"
	"golang-web/internal/model"

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

func (r *RecipesRepository) ListSummaries(tx *sql.Tx, recipeType string) ([]model.RecipeSummary, error) {
	query := `
		SELECT r.id, r.name, r.type, r.yield_qty, r.yield_unit, COUNT(ri.id) AS item_count 
		FROM recipes r 
		LEFT JOIN recipe_items ri ON ri.recipe_id = r.id 
		WHERE ($1 = '' OR r.type = $1) 
		GROUP BY r.id 
		ORDER BY r.name 
	`

	var recipes []model.RecipeSummary

	rows, err := tx.Query(query, recipeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		recipe := model.RecipeSummary{}
		if err := rows.Scan(
			&recipe.ID,
			&recipe.Name,
			&recipe.Type,
			&recipe.YieldQty,
			&recipe.YieldUnit,
			&recipe.ItemCount,
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
