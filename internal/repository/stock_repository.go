package repository

import (
	"database/sql"
	"golang-web/internal/model"
	"time"

	"github.com/sirupsen/logrus"
)

type MovementFilter struct {
	IngredientID, Type, From, To string
}

type StockRepository struct {
	Log *logrus.Logger
}

func NewStockRepository(logger *logrus.Logger) *StockRepository {
	return &StockRepository{
		Log: logger,
	}
}

func (r *StockRepository) ListStocks(tx *sql.Tx, q string) ([]model.StockRecord, error) {
	rows, e := tx.Query(`SELECT i.id,i.name,i.base_unit,i.purchase_unit,i.base_qty_per_purchase_unit::text,COALESCE(SUM(m.quantity),0)::text,MAX(m.movement_date) FROM ingredients i LEFT JOIN stock_movements m ON m.ingredient_id=i.id WHERE ($1::text='' OR strpos(lower(i.name),lower($1::text))>0) GROUP BY i.id ORDER BY i.name`, q)
	if e != nil {
		return nil, e
	}
	defer rows.Close()

	out := []model.StockRecord{}
	for rows.Next() {
		var v model.StockRecord
		var d sql.NullTime
		if e = rows.Scan(&v.IngredientID, &v.Name, &v.BaseUnit, &v.PurchaseUnit, &v.BaseQtyPerPurchaseUnit, &v.StockQty, &d); e != nil {
			return nil, e
		}
		if d.Valid {
			v.LastMovement = &d.Time
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

func (r *StockRepository) ListIngredientOptions(tx *sql.Tx) ([]model.IngredientOption, error) {
	rows, e := tx.Query(`SELECT i.id,i.name,i.base_unit,i.purchase_unit,i.base_qty_per_purchase_unit::text,COALESCE(SUM(m.quantity),0)::text FROM ingredients i LEFT JOIN stock_movements m ON m.ingredient_id=i.id GROUP BY i.id ORDER BY i.name`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()

	out := []model.IngredientOption{}
	for rows.Next() {
		var v model.IngredientOption
		if e = rows.Scan(&v.ID, &v.Name, &v.BaseUnit, &v.PurchaseUnit, &v.BaseQtyPerPurchaseUnit, &v.StockQty); e != nil {
			return nil, e
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

func (r *StockRepository) LockIngredient(tx *sql.Tx, id int64) (*model.IngredientOption, error) {
	v := new(model.IngredientOption)
	e := tx.QueryRow(`SELECT id,name,base_unit,purchase_unit,base_qty_per_purchase_unit::text FROM ingredients WHERE id=$1 FOR UPDATE`, id).Scan(&v.ID, &v.Name, &v.BaseUnit, &v.PurchaseUnit, &v.BaseQtyPerPurchaseUnit)

	return v, e
}

func (r *StockRepository) ProjectedStock(tx *sql.Tx, id int64, delta string) (string, bool, error) {
	var s string
	var n bool
	e := tx.QueryRow(`SELECT (COALESCE(SUM(quantity),0)+$2::numeric)::text,(COALESCE(SUM(quantity),0)+$2::numeric)<0 FROM stock_movements WHERE ingredient_id=$1`, id, delta).Scan(&s, &n)
	return s, n, e
}

func (r *StockRepository) InsertPurchase(tx *sql.Tx, id int64, d time.Time, q, c, n string) error {
	_, e := tx.Exec(`INSERT INTO stock_movements (ingredient_id,movement_date,type,quantity,total_cost,note) SELECT i.id,$2::date,'PURCHASE',$3::numeric*i.base_qty_per_purchase_unit,$4::numeric,NULLIF($5::text,'') FROM ingredients i WHERE i.id=$1`, id, d, q, c, n)
	return e
}

func (r *StockRepository) InsertUsage(tx *sql.Tx, id int64, d time.Time, q string, p time.Time, n string) error {
	_, e := tx.Exec(`INSERT INTO stock_movements (ingredient_id,movement_date,type,quantity,production_date,note) VALUES ($1,$2::date,'USAGE',$3::numeric,$4::date,NULLIF($5::text,''))`, id, d, q, p, n)
	return e
}

func (r *StockRepository) InsertAdjustment(tx *sql.Tx, id int64, d time.Time, q, n string) error {
	_, e := tx.Exec(`INSERT INTO stock_movements (ingredient_id,movement_date,type,quantity,note) VALUES ($1,$2::date,'ADJUSTMENT',$3::numeric,NULLIF($4::text,''))`, id, d, q, n)
	return e
}

func (r *StockRepository) PriceExistsOnOrAfter(tx *sql.Tx, id int64, d time.Time) (bool, error) {
	var x bool
	e := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM ingredient_prices WHERE ingredient_id=$1 AND effective_from >= $2::date)`, id, d).Scan(&x)
	return x, e
}

func (r *StockRepository) InsertPrice(tx *sql.Tx, id int64, c, q string, d time.Time) error {
	_, e := tx.Exec(`INSERT INTO ingredient_prices (ingredient_id,price_per_purchase_unit,effective_from) VALUES ($1,ROUND($2::numeric/$3::numeric,2),$4::date)`, id, c, q, d)
	return e
}

func movementWhere(f MovementFilter) string {
	return ` WHERE ($1::text='' OR m.ingredient_id=$1::int) AND ($2::text='' OR m.type=$2::text) AND ($3::text='' OR m.movement_date >= $3::date) AND ($4::text='' OR m.movement_date <= $4::date)`
}

func (r *StockRepository) ListMovements(tx *sql.Tx, f MovementFilter, l, o int) ([]model.MovementRecord, error) {
	q := `SELECT m.movement_date,i.name,m.type,m.quantity::text,i.base_unit,i.purchase_unit,i.base_qty_per_purchase_unit::text,m.production_date,m.total_cost::text,COALESCE(m.note,'') FROM stock_movements m JOIN ingredients i ON i.id=m.ingredient_id` + movementWhere(f) + ` ORDER BY m.movement_date DESC,m.id DESC LIMIT $5 OFFSET $6`
	rows, e := tx.Query(q, f.IngredientID, f.Type, f.From, f.To, l, o)
	if e != nil {
		return nil, e
	}
	defer rows.Close()

	out := []model.MovementRecord{}
	for rows.Next() {
		var v model.MovementRecord
		var pd sql.NullTime
		var cost sql.NullString
		if e = rows.Scan(&v.Date, &v.IngredientName, &v.Type, &v.Quantity, &v.BaseUnit, &v.PurchaseUnit, &v.BaseQtyPerPurchaseUnit, &pd, &cost, &v.Note); e != nil {
			return nil, e
		}
		if pd.Valid {
			v.ProductionDate = &pd.Time
		}
		if cost.Valid {
			x := cost.String
			v.TotalCost = &x
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

func (r *StockRepository) CountMovements(tx *sql.Tx, f MovementFilter) (int, error) {
	var n int
	e := tx.QueryRow(`SELECT COUNT(*) FROM stock_movements m`+movementWhere(f), f.IngredientID, f.Type, f.From, f.To).Scan(&n)
	return n, e
}
