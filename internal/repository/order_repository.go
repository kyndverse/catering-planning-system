package repository

import (
	"database/sql"
	"time"

	"golang-web/internal/model"

	"github.com/sirupsen/logrus"
)

type OrderRepository struct {
	Log *logrus.Logger
}

func NewOrderRepository(logger *logrus.Logger) *OrderRepository {
	return &OrderRepository{
		Log: logger,
	}
}

func (r *OrderRepository) ListSummaries(tx *sql.Tx, eventDate, productionDate *time.Time) ([]model.OrderSummary, error) {
	query := `
				SELECT o.id,o.customer_name,o.event_date,o.production_date,COUNT(oi.menu_id),COALESCE(SUM(oi.portions),0)
				FROM orders o LEFT JOIN order_items oi ON oi.order_id=o.id
				WHERE ($1::date IS NULL OR o.event_date=$1::date) AND ($2::date IS NULL OR o.production_date=$2::date)
				GROUP BY o.id ORDER BY o.production_date,o.event_date,o.id	
	`

	rows, err := tx.Query(query, eventDate, productionDate)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []model.OrderSummary
	for rows.Next() {
		var v model.OrderSummary
		if err := rows.Scan(&v.ID, &v.CustomerName, &v.EventDate, &v.ProductionDate, &v.MenuCount, &v.TotalPortions); err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

func (r *OrderRepository) FindOne(tx *sql.Tx, id int64) (*model.OrderDetail, error) {
	v := new(model.OrderDetail)
	err := tx.QueryRow(`SELECT o.id,o.customer_name,o.event_date,o.production_date,COALESCE((SELECT SUM(portions) FROM order_items WHERE order_id=o.id),0) FROM orders o WHERE o.id=$1`, id).Scan(&v.ID, &v.CustomerName, &v.EventDate, &v.ProductionDate, &v.TotalPortions)
	return v, err
}

func (r *OrderRepository) Create(tx *sql.Tx, name string, event, production time.Time) (int64, error) {
	var id int64
	err := tx.QueryRow(`INSERT INTO orders(customer_name,event_date,production_date) VALUES($1,$2,$3) RETURNING id`, name, event, production).Scan(&id)
	return id, err
}

func (r *OrderRepository) Update(tx *sql.Tx, id int64, name string, event, production time.Time) (bool, error) {
	res, err := tx.Exec(`UPDATE orders SET customer_name=$2,event_date=$3,production_date=$4 WHERE id=$1`, id, name, event, production)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *OrderRepository) Delete(tx *sql.Tx, id int64) (bool, error) {
	res, err := tx.Exec(`DELETE FROM orders WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *OrderRepository) ListItemsByOrderID(tx *sql.Tx, id int64) ([]model.OrderItemView, error) {
	rows, err := tx.Query(`SELECT oi.menu_id,r.name,oi.portions FROM order_items oi JOIN recipes r ON r.id=oi.menu_id WHERE oi.order_id=$1 ORDER BY r.name`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.OrderItemView
	for rows.Next() {
		var v model.OrderItemView
		if err := rows.Scan(&v.MenuID, &v.MenuName, &v.Portions); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *OrderRepository) ListMenuOptions(tx *sql.Tx, id int64) ([]model.MenuOption, error) {
	rows, err := tx.Query(`SELECT r.id,r.name FROM recipes r WHERE r.type='MENU' AND NOT EXISTS(SELECT 1 FROM order_items oi WHERE oi.order_id=$1 AND oi.menu_id=r.id) ORDER BY r.name`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.MenuOption
	for rows.Next() {
		var v model.MenuOption
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *OrderRepository) AddItem(tx *sql.Tx, orderID, menuID int64, portions int) error {
	_, err := tx.Exec(`INSERT INTO order_items(order_id,menu_id,portions) VALUES($1,$2,$3)`, orderID, menuID, portions)
	return err
}

func (r *OrderRepository) UpdateItemPortions(tx *sql.Tx, orderID, menuID int64, portions int) (bool, error) {
	res, err := tx.Exec(`UPDATE order_items SET portions=$3 WHERE order_id=$1 AND menu_id=$2`, orderID, menuID, portions)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *OrderRepository) DeleteItem(tx *sql.Tx, orderID, menuID int64) (bool, error) {
	res, err := tx.Exec(`DELETE FROM order_items WHERE order_id=$1 AND menu_id=$2`, orderID, menuID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *OrderRepository) OrderExists(tx *sql.Tx, id int64) (bool, error) {
	var v bool
	err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM orders WHERE id=$1)`, id).Scan(&v)
	return v, err
}

func (r *OrderRepository) FindRecipeType(tx *sql.Tx, id int64) (string, error) {
	var v string
	err := tx.QueryRow(`SELECT type FROM recipes WHERE id=$1`, id).Scan(&v)
	return v, err
}

func (r *OrderRepository) ItemExists(tx *sql.Tx, orderID, menuID int64) (bool, error) {
	var v bool
	err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM order_items WHERE order_id=$1 AND menu_id=$2)`, orderID, menuID).Scan(&v)
	return v, err
}
