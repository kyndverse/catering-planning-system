package repository

import (
	"database/sql"
	"fmt"
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

func (r *RecipeRepository) RecipeExists(tx *sql.Tx, id int64) (bool, error) {
	var exists bool
	err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM recipes WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *RecipeRepository) IngredientExists(tx *sql.Tx, id int64) (bool, error) {
	var exists bool
	err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM ingredients WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

// FindType mengembalikan jenis resep (PREPARED/COMPONENT/MENU).
// Mengembalikan sql.ErrNoRows jika resep tidak ada.
func (r *RecipeRepository) FindType(tx *sql.Tx, id int64) (string, error) {
	var recipeType string
	err := tx.QueryRow(`SELECT type FROM recipes WHERE id = $1`, id).Scan(&recipeType)
	return recipeType, err
}

// WouldCreateCycle: true jika menjadikan subRecipeID sebagai isi recipeID akan
// membuat siklus, yaitu subRecipeID ternyata leluhur recipeID (langsung maupun
// tidak langsung memuat recipeID). Kasus subRecipeID == recipeID dicek terpisah.
func (r *RecipeRepository) WouldCreateCycle(tx *sql.Tx, recipeID, subRecipeID int64) (bool, error) {
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT recipe_id FROM recipe_items WHERE sub_recipe_id = $1
			UNION
			SELECT ri.recipe_id
			FROM recipe_items ri
			JOIN ancestors a ON ri.sub_recipe_id = a.recipe_id
		)
		SELECT EXISTS (SELECT 1 FROM ancestors WHERE recipe_id = $2)`

	var cycle bool
	err := tx.QueryRow(query, recipeID, subRecipeID).Scan(&cycle)
	return cycle, err
}

// ItemExists: true jika bahan/sub-resep itu sudah menjadi isi resep ini.
func (r *RecipeRepository) ItemExists(tx *sql.Tx, recipeID int64, kind string, targetID int64) (bool, error) {
	var query string
	switch kind {
	case model.ItemKindIngredient:
		query = `SELECT EXISTS (SELECT 1 FROM recipe_items WHERE recipe_id = $1 AND ingredient_id = $2)`
	case model.ItemKindRecipe:
		query = `SELECT EXISTS (SELECT 1 FROM recipe_items WHERE recipe_id = $1 AND sub_recipe_id = $2)`
	default:
		return false, fmt.Errorf("unknown item kind %q", kind)
	}

	var exists bool
	err := tx.QueryRow(query, recipeID, targetID).Scan(&exists)
	return exists, err
}

// AddItem menambah satu baris recipe_items. quantity berupa string desimal
// yang sudah divalidasi; dikirim apa adanya ke NUMERIC (tanpa float).
func (r *RecipeRepository) AddItem(tx *sql.Tx, recipeID int64, kind string, targetID int64, quantity string) error {
	var ingredientID, subRecipeID sql.NullInt64

	switch kind {
	case model.ItemKindIngredient:
		ingredientID = sql.NullInt64{Int64: targetID, Valid: true}
	case model.ItemKindRecipe:
		subRecipeID = sql.NullInt64{Int64: targetID, Valid: true}
	default:
		return fmt.Errorf("unknown item kind %q", kind)
	}

	query := `
		INSERT INTO recipe_items (recipe_id, ingredient_id, sub_recipe_id, quantity)
		VALUES ($1, $2, $3, $4::numeric)`

	_, err := tx.Exec(query, recipeID, ingredientID, subRecipeID, quantity)
	return err
}

func (r *RecipeRepository) UpdateItemQuantity(tx *sql.Tx, recipeID, itemID int64, quantity string) (bool, error) {
	query := `UPDATE recipe_items SET quantity = $3::numeric WHERE id = $1 AND recipe_id = $2`

	res, err := tx.Exec(query, itemID, recipeID, quantity)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// DeleteItem menghapus satu isi dari resep. Mengembalikan false jika tidak ada
// baris yang cocok (item tidak ada, atau bukan milik resep di URL).
func (r *RecipeRepository) DeleteItem(tx *sql.Tx, recipeID, itemID int64) (bool, error) {
	query := `DELETE FROM recipe_items WHERE id = $1 AND recipe_id = $2`

	res, err := tx.Exec(query, itemID, recipeID)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}