package model

import "time"

type SearchStockRequest struct{ Q string }
type SearchMovementRequest struct{ IngredientID, Type, From, To, Page string }
type CreatePurchaseRequest struct {
	MovementDate string `form:"movement_date" validate:"required"`
	IngredientID string `form:"ingredient_id" validate:"required"`
	Quantity     string `form:"quantity" validate:"required"`
	TotalCost    string `form:"total_cost" validate:"required"`
	Note         string `form:"note"`
	UpdatePrice  string `form:"update_price"`
}
type CreateUsageRequest struct {
	MovementDate    string `form:"movement_date" validate:"required"`
	ProductionDate  string `form:"production_date" validate:"required"`
	IngredientID    string `form:"ingredient_id" validate:"required"`
	Quantity        string `form:"quantity" validate:"required"`
	Note            string `form:"note"`
	ConfirmNegative string `form:"confirm_negative"`
}
type CreateAdjustmentRequest struct {
	MovementDate    string `form:"movement_date" validate:"required"`
	IngredientID    string `form:"ingredient_id" validate:"required"`
	Direction       string `form:"direction" validate:"required"`
	Quantity        string `form:"quantity" validate:"required"`
	Note            string `form:"note" validate:"required"`
	ConfirmNegative string `form:"confirm_negative"`
}
type StockRecord struct {
	IngredientID                                                   int64
	Name, BaseUnit, PurchaseUnit, BaseQtyPerPurchaseUnit, StockQty string
	LastMovement                                                   *time.Time
}
type IngredientOption struct {
	ID                                                             int64
	Name, BaseUnit, PurchaseUnit, BaseQtyPerPurchaseUnit, StockQty string
}
type MovementRecord struct {
	Date                                                                           time.Time
	IngredientName, Type, Quantity, BaseUnit, PurchaseUnit, BaseQtyPerPurchaseUnit string
	ProductionDate                                                                 *time.Time
	TotalCost                                                                      *string
	Note                                                                           string
}
type StockWriteResult struct {
	IngredientID int64
	Warning      string
}
