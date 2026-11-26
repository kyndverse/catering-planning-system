package repository

import (
	"database/sql"
	"golang-web/internal/model"
	"time"
)

type UsageReportRepository struct{}

func NewUsageReportRepository() *UsageReportRepository { return &UsageReportRepository{} }
func (r *UsageReportRepository) ListUsageDates(tx *sql.Tx, today time.Time) ([]model.UsageDateSummary, error) {
	rows, err := tx.Query(`WITH d AS (SELECT production_date d FROM orders WHERE production_date <= $1::date UNION SELECT production_date FROM stock_movements WHERE type='USAGE' AND production_date <= $1::date) SELECT d.d,(SELECT COUNT(*) FROM orders o WHERE o.production_date=d.d),(SELECT COALESCE(SUM(oi.portions),0) FROM orders o JOIN order_items oi ON oi.order_id=o.id WHERE o.production_date=d.d),EXISTS(SELECT 1 FROM stock_movements m WHERE m.type='USAGE' AND m.production_date=d.d) FROM d ORDER BY d.d DESC LIMIT 30`, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.UsageDateSummary
	for rows.Next() {
		var x model.UsageDateSummary
		if err := rows.Scan(&x.Date, &x.OrderCount, &x.TotalPortions, &x.UsageRecorded); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *UsageReportRepository) ListPlanVsActual(tx *sql.Tx, date, today time.Time) ([]model.UsageInputRow, error) {
	rows, err := tx.Query(`WITH RECURSIVE menu_totals AS (SELECT oi.menu_id recipe_id,SUM(oi.portions)::numeric qty FROM orders o JOIN order_items oi ON oi.order_id=o.id WHERE o.production_date=$1::date GROUP BY oi.menu_id), exploded AS (SELECT recipe_id,qty,0 depth FROM menu_totals UNION ALL SELECT ri.sub_recipe_id,e.qty*ri.quantity/r.yield_qty,e.depth+1 FROM exploded e JOIN recipes r ON r.id=e.recipe_id JOIN recipe_items ri ON ri.recipe_id=r.id AND ri.sub_recipe_id IS NOT NULL WHERE e.depth<20), planned AS (SELECT ri.ingredient_id,ROUND(SUM(e.qty*ri.quantity/r.yield_qty),4) planned_qty FROM exploded e JOIN recipes r ON r.id=e.recipe_id JOIN recipe_items ri ON ri.recipe_id=r.id AND ri.ingredient_id IS NOT NULL GROUP BY ri.ingredient_id), actual AS (SELECT ingredient_id,-SUM(quantity) actual_qty FROM stock_movements WHERE type='USAGE' AND production_date=$1::date GROUP BY ingredient_id), latest_price AS (SELECT DISTINCT ON (ingredient_id) ingredient_id,price_per_purchase_unit FROM ingredient_prices WHERE effective_from <= $2::date ORDER BY ingredient_id,effective_from DESC) SELECT i.id,i.name,i.base_unit,i.base_qty_per_purchase_unit::text,COALESCE(p.planned_qty,0)::text,COALESCE(a.actual_qty,0)::text,(a.ingredient_id IS NOT NULL),lp.price_per_purchase_unit::text FROM planned p FULL OUTER JOIN actual a ON a.ingredient_id=p.ingredient_id JOIN ingredients i ON i.id=COALESCE(p.ingredient_id,a.ingredient_id) LEFT JOIN latest_price lp ON lp.ingredient_id=i.id ORDER BY i.name`, date, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.UsageInputRow
	for rows.Next() {
		var x model.UsageInputRow
		var p sql.NullString
		if err := rows.Scan(&x.IngredientID, &x.IngredientName, &x.BaseUnit, &x.BaseQtyPerPurchaseUnit, &x.PlannedQty, &x.ActualQty, &x.HasActual, &p); err != nil {
			return nil, err
		}
		if p.Valid {
			v := p.String
			x.PricePerPurchaseUnit = &v
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
