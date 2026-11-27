package repository

import (
	"database/sql"
	"time"

	"golang-web/internal/model"

	"github.com/sirupsen/logrus"
)

type ProductionRepository struct {
	Log *logrus.Logger
}

func NewProductionRepository(logger *logrus.Logger) *ProductionRepository {
	return &ProductionRepository{
		Log: logger,
	}
}

func (r *ProductionRepository) ListUpcomingDates(tx *sql.Tx, today time.Time) ([]model.ProductionDateSummary, error) {
	query := `
				SELECT o.production_date, COUNT(DISTINCT o.id), COALESCE(SUM(oi.portions), 0), MAX(o.event_date)
				FROM orders o LEFT JOIN order_items oi ON oi.order_id = o.id
				WHERE o.production_date >= $1::date GROUP BY o.production_date ORDER BY o.production_date LIMIT 60
	`

	rows, err := tx.Query(query, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ProductionDateSummary
	for rows.Next() {
		var row model.ProductionDateSummary
		if err := rows.Scan(&row.Date, &row.OrderCount, &row.TotalPortions, &row.MaxEventDate); err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	return out, rows.Err()
}

func (r *ProductionRepository) ListMenuTotals(tx *sql.Tx, productionDate time.Time) ([]model.ProductionMenu, error) {
	rows, err := tx.Query(`SELECT oi.menu_id, r.name, SUM(oi.portions) FROM orders o JOIN order_items oi ON oi.order_id=o.id JOIN recipes r ON r.id=oi.menu_id WHERE o.production_date=$1::date GROUP BY oi.menu_id,r.name ORDER BY r.name`, productionDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ProductionMenu
	for rows.Next() {
		var row model.ProductionMenu
		if err := rows.Scan(&row.MenuID, &row.MenuName, &row.Portions); err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	return out, rows.Err()
}

func (r *ProductionRepository) ListRequirements(tx *sql.Tx, productionDate, today time.Time) ([]model.ProductionRequirementRow, error) {
	query := `
				WITH RECURSIVE menu_totals AS (
				SELECT oi.menu_id AS recipe_id, SUM(oi.portions)::numeric AS qty FROM orders o JOIN order_items oi ON oi.order_id=o.id WHERE o.production_date=$1::date GROUP BY oi.menu_id
				), exploded AS (
				SELECT recipe_id, qty, 0 AS depth FROM menu_totals
				UNION ALL
				SELECT ri.sub_recipe_id, e.qty * ri.quantity / r.yield_qty, e.depth + 1 FROM exploded e JOIN recipes r ON r.id=e.recipe_id JOIN recipe_items ri ON ri.recipe_id=r.id AND ri.sub_recipe_id IS NOT NULL WHERE e.depth < 20
				), requirements AS (
				SELECT ri.ingredient_id, ROUND(SUM(e.qty * ri.quantity / r.yield_qty), 4) AS required_qty FROM exploded e JOIN recipes r ON r.id=e.recipe_id JOIN recipe_items ri ON ri.recipe_id=r.id AND ri.ingredient_id IS NOT NULL GROUP BY ri.ingredient_id
				), stock AS (SELECT ingredient_id,SUM(quantity) AS stock_qty FROM stock_movements GROUP BY ingredient_id), latest_price AS (
				SELECT DISTINCT ON (ip.ingredient_id) ip.ingredient_id,ip.price_per_purchase_unit FROM ingredient_prices ip WHERE ip.effective_from <= $2::date ORDER BY ip.ingredient_id,ip.effective_from DESC
				)
				SELECT i.id,i.name,i.base_unit,i.purchase_unit,i.base_qty_per_purchase_unit::text,q.required_qty::text,COALESCE(s.stock_qty,0)::text,p.price_per_purchase_unit::text FROM requirements q JOIN ingredients i ON i.id=q.ingredient_id LEFT JOIN stock s ON s.ingredient_id=q.ingredient_id LEFT JOIN latest_price p ON p.ingredient_id=q.ingredient_id ORDER BY i.name
	`

	rows, err := tx.Query(query, productionDate, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ProductionRequirementRow
	for rows.Next() {
		var row model.ProductionRequirementRow
		var price sql.NullString
		if err := rows.Scan(&row.IngredientID, &row.Name, &row.BaseUnit, &row.PurchaseUnit, &row.BaseQtyPerPurchaseUnit, &row.RequiredQty, &row.StockQty, &price); err != nil {
			return nil, err
		}
		if price.Valid {
			value := price.String
			row.PricePerPurchaseUnit = &value
		}
		out = append(out, row)
	}

	return out, rows.Err()
}
