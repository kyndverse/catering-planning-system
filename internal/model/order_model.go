package model

import "time"

type SearchOrderRequest struct {
	EventDate      string
	ProductionDate string
}

type CreateOrderRequest struct {
	CustomerName   string `form:"customer_name" validate:"required"`
	EventDate      string `form:"event_date" validate:"required"`
	ProductionDate string `form:"production_date"`
}

type UpdateOrderRequest struct {
	ID             int64  `form:"-"`
	CustomerName   string `form:"customer_name" validate:"required"`
	EventDate      string `form:"event_date" validate:"required"`
	ProductionDate string `form:"production_date"`
}

type DeleteOrderRequest struct{ ID int64 }

type AddOrderItemRequest struct {
	OrderID  int64  `form:"-"`
	MenuID   string `form:"menu_id" validate:"required"`
	Portions string `form:"portions" validate:"required"`
}

type UpdateOrderItemRequest struct {
	OrderID  int64  `form:"-"`
	MenuID   int64  `form:"-"`
	Portions string `form:"portions" validate:"required"`
}

type DeleteOrderItemRequest struct {
	OrderID int64
	MenuID  int64
}

type OrderSummary struct {
	ID             int64
	CustomerName   string
	EventDate      time.Time
	ProductionDate time.Time
	MenuCount      int
	TotalPortions  int
	Locked         bool
}

type OrderDetail struct {
	ID             int64
	CustomerName   string
	EventDate      time.Time
	ProductionDate time.Time
	TotalPortions  int
	Locked         bool
}

type OrderItemView struct {
	MenuID   int64
	MenuName string
	Portions int
}

type MenuOption struct {
	ID   int64
	Name string
}
