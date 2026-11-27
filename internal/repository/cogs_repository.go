package repository

import (
	"database/sql"
	"golang-web/internal/model"
	"time"

	"github.com/sirupsen/logrus"
)

type CogsRepository struct {
	Log *logrus.Logger
}

func NewCogsRepository(logger *logrus.Logger) *CogsRepository {
	return &CogsRepository{
		Log: logger,
	}
}

func (r *CogsRepository) ListMenus(tx *sql.Tx) ([]model.MenuRef, error) {
	query := `SELECT id,name FROM recipes WHERE type='MENU' ORDER BY name`

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.MenuRef
	for rows.Next() {
		var x model.MenuRef
		if err := rows.Scan(&x.ID, &x.Name); err != nil {
			return nil, err
		}

		out = append(out, x)
	}

	return out, rows.Err()
}

func (r *CogsRepository) FindMenu(tx *sql.Tx, id int64) (*model.MenuRef, error) {
	query := `SELECT id,name FROM recipes WHERE id=$1 AND type='MENU'`

	x := new(model.MenuRef)
	err := tx.QueryRow(query, id).Scan(&x.ID, &x.Name)
	return x, err
}

func (r *CogsRepository) ListMenuIngredientCosts(tx *sql.Tx, menuID string, today time.Time) ([]model.MenuIngredientCostRow, error) {
	query := `
        WITH RECURSIVE exploded AS (
				SELECT r.id menu_id,r.id recipe_id,1::numeric qty,0 depth FROM recipes r WHERE r.type='MENU' AND ($1::text='' OR r.id=$1::int)
				UNION ALL SELECT e.menu_id,ri.sub_recipe_id,e.qty*ri.quantity/r.yield_qty,e.depth+1 FROM exploded e JOIN recipes r ON r.id=e.recipe_id JOIN recipe_items ri ON ri.recipe_id=r.id AND ri.sub_recipe_id IS NOT NULL WHERE e.depth<20
				), menu_ingredients AS (SELECT e.menu_id,ri.ingredient_id,ROUND(SUM(e.qty*ri.quantity/r.yield_qty),6) qty_per_portion FROM exploded e JOIN recipes r ON r.id=e.recipe_id JOIN recipe_items ri ON ri.recipe_id=r.id AND ri.ingredient_id IS NOT NULL GROUP BY e.menu_id,ri.ingredient_id), latest_price AS (SELECT DISTINCT ON (ingredient_id) ingredient_id,price_per_purchase_unit,effective_from FROM ingredient_prices WHERE effective_from <= $2::date ORDER BY ingredient_id,effective_from DESC)
				SELECT mi.menu_id,i.id,i.name,i.base_unit,i.base_qty_per_purchase_unit::text,mi.qty_per_portion::text,p.price_per_purchase_unit::text,p.effective_from FROM menu_ingredients mi JOIN ingredients i ON i.id=mi.ingredient_id LEFT JOIN latest_price p ON p.ingredient_id=mi.ingredient_id ORDER BY mi.menu_id,i.name
    `

	rows, err := tx.Query(query, menuID, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.MenuIngredientCostRow
	for rows.Next() {
		var x model.MenuIngredientCostRow
		var p sql.NullString
		var since sql.NullTime
		if err := rows.Scan(&x.MenuID, &x.IngredientID, &x.IngredientName, &x.BaseUnit, &x.BaseQtyPerPurchaseUnit, &x.QtyPerPortion, &p, &since); err != nil {
			return nil, err
		}
		if p.Valid {
			v := p.String
			x.PricePerPurchaseUnit = &v
		}

		if since.Valid {
			v := since.Time
			x.PriceSince = &v
		}

		out = append(out, x)
	}

	return out, rows.Err()
}
