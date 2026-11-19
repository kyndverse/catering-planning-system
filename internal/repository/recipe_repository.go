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

func (r *RecipeRepository) ListItemsByRecipeID(tx *sql.Tx, recipeID int64) ([]model.RecipeItemView, error) {
	query :=
		`
			SELECT ri.id,
					CASE WHEN ri.sub_recipe_id IS NOT NULL THEN 'recipe' ELSE 'ingredient' END AS kind,
					COALESCE(ri.sub_recipe_id, ri.ingredient_id) AS target_id,
					COALESCE(sr.name, i.name) AS name,
					ri.quantity::text AS quantity,
					COALESCE(sr.yield_unit, i.base_unit) AS unit
			FROM recipe_items ri
			LEFT JOIN ingredients i ON i.id  = ri.ingredient_id
			LEFT JOIN recipes sr    ON sr.id = ri.sub_recipe_id
			WHERE ri.recipe_id = $1
			ORDER BY ri.id;
		`

	rows, err := tx.Query(query, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.RecipeItemView
	for rows.Next() {
		var it model.RecipeItemView
		if err := rows.Scan(&it.ID, &it.Kind, &it.TargetID, &it.Name, &it.Quantity, &it.Unit); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *RecipeRepository) ListSubRecipeOptions(tx *sql.Tx, ID int64) ([]model.SubRecipeItemOption, error) {
	query :=
		`
			WITH RECURSIVE ancestors AS (
				SELECT recipe_id FROM recipe_items WHERE sub_recipe_id = $1
				UNION
				SELECT ri.recipe_id
				FROM recipe_items ri
				JOIN ancestors a ON ri.sub_recipe_id = a.recipe_id
			)
			SELECT r.id, r.name, r.yield_unit AS unit
			FROM recipes r
			WHERE r.id <> $1
				AND r.type <> 'MENU'
				AND r.id NOT IN (SELECT recipe_id FROM ancestors)
			ORDER BY r.name;
		`

	rows, err := tx.Query(query, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.SubRecipeItemOption
	for rows.Next() {
		var it model.SubRecipeItemOption
		if err := rows.Scan(&it.ID, &it.Name, &it.Unit); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *RecipeRepository) ExplodeToIngredients(tx *sql.Tx, recipeID int64, qty string) ([]model.IngredientRequirement, error) {
	query := `
			WITH RECURSIVE exploded AS (
				SELECT r.id AS recipe_id, $2::numeric AS qty, 0 AS depth
				FROM recipes r
				WHERE r.id = $1

				UNION ALL

				SELECT ri.sub_recipe_id, e.qty / r.yield_qty * ri.quantity, e.depth + 1
				FROM exploded e
				JOIN recipes r       ON r.id = e.recipe_id
				JOIN recipe_items ri ON ri.recipe_id = r.id AND ri.sub_recipe_id IS NOT NULL
				WHERE e.depth < 20
			)
			SELECT i.name,
						SUM(e.qty / r.yield_qty * ri.quantity)::text AS quantity,
						i.base_unit
			FROM exploded e
			JOIN recipes r       ON r.id = e.recipe_id
			JOIN recipe_items ri ON ri.recipe_id = r.id AND ri.ingredient_id IS NOT NULL
			JOIN ingredients i   ON i.id = ri.ingredient_id
			GROUP BY i.id, i.name, i.base_unit
			ORDER BY i.name`

	rows, err := tx.Query(query, recipeID, qty)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []model.IngredientRequirement
	for rows.Next() {
		var q model.IngredientRequirement
		if err := rows.Scan(&q.IngredientName, &q.Quantity, &q.Unit); err != nil {
			return nil, err
		}
		reqs = append(reqs, q)
	}
	return reqs, rows.Err()
}
