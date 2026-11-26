package model

import (
	"math/big"
	"time"
)

type FormData[T any] struct {
	Judul  string
	Action string
	Data   T
	Errors map[string]string
}

type MenuRef struct {
	ID   int64
	Name string
}
type MenuIngredientCostRow struct {
	MenuID, IngredientID                                            int64
	IngredientName, BaseUnit, BaseQtyPerPurchaseUnit, QtyPerPortion string
	PricePerPurchaseUnit                                            *string
	PriceSince                                                      *time.Time
}
type CostLine struct {
	IngredientName, Unit    string
	Qty, PricePerBase, Cost *big.Rat
	PriceSince              *time.Time
}
type MenuCogs struct {
	Menu    MenuRef
	Lines   []CostLine
	Total   *big.Rat
	Missing int
	Status  string
}

type UsageDateSummary struct {
	Date                      time.Time
	OrderCount, TotalPortions int
	UsageRecorded             bool
}
type UsageInputRow struct {
	IngredientID                                                            int64
	IngredientName, BaseUnit, BaseQtyPerPurchaseUnit, PlannedQty, ActualQty string
	HasActual                                                               bool
	PricePerPurchaseUnit                                                    *string
}
type UsageLine struct {
	IngredientName, Unit, Status          string
	Planned, Actual, Variance, Pct, Value *big.Rat
}
type UsageSummary struct {
	OverCount, UnderCount, UnrecordedCount int
	TotalValue                             *big.Rat
	HasValue, PriceIncomplete              bool
}
type UsageReport struct {
	Date                      time.Time
	OrderCount, TotalPortions int
	Lines                     []UsageLine
	Summary                   UsageSummary
}
