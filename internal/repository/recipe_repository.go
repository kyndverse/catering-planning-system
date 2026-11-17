package repository

import (
	"database/sql"
	"golang-web/internal/domain/entity"
	"golang-web/internal/model"

	"github.com/sirupsen/logrus"
)

type RecipeRepository struct {
	Log *logrus.Logger
}

func NewRecipesRepository(logger *logrus.Logger) *RecipeRepository {
	return &RecipeRepository{
		Log: logger,
	}
}

func (r *RecipeRepository) ListSummaries(tx *sql.Tx, recipeType string) ([]model.RecipeSummary, error) {
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

func (r *RecipeRepository) Create(tx *sql.Tx, recipe *entity.Recipe) error {
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

func (r *RecipeRepository) Update(tx *sql.Tx, recipe *entity.Recipe) error {
	query := `
		UPDATE recipes
		SET name=$1, type=$2, yield_qty=$3, yield_unit=$4
		WHERE id=$5
		RETURNING id, name, type, yield_qty, yield_unit
	`

	return tx.QueryRow(
		query,
		recipe.Name,
		recipe.Type,
		recipe.YieldQty,
		recipe.YieldUnit,
		recipe.ID,
	).Scan(
		&recipe.ID,
		&recipe.Name,
		&recipe.Type,
		&recipe.YieldQty,
		&recipe.YieldUnit)
}

func (r *RecipeRepository) Delete(tx *sql.Tx, recipe *entity.Recipe) error {
	query := `DELETE FROM recipes WHERE id=$1`

	_, err := tx.Exec(query, recipe.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *RecipeRepository) FindOne(tx *sql.Tx, recipe *entity.Recipe) error {
	query := `SELECT id, name, type, yield_qty, yield_unit FROM recipes WHERE id=$1 LIMIT 1`

	return tx.QueryRow(
		query,
		recipe.ID,
	).Scan(
		&recipe.ID,
		&recipe.Name,
		&recipe.Type,
		&recipe.YieldQty,
		&recipe.YieldUnit,
	)
}
