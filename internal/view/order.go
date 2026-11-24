package view

import (
	"golang-web/internal/model"
	"golang-web/internal/timeutil"
)

type OrderSummary struct {
	ID                                      int64
	CustomerName, EventDate, ProductionDate string
	MenuCount, TotalPortions                int
	Locked                                  bool
}
type OrderDetail struct {
	ID                                      int64
	CustomerName, EventDate, ProductionDate string
	TotalPortions                           int
	Locked                                  bool
}

func OrderSummaries(rows []model.OrderSummary) []OrderSummary {
	out := make([]OrderSummary, 0, len(rows))
	for _, v := range rows {
		out = append(out, OrderSummary{v.ID, v.CustomerName, timeutil.FormatDateID(v.EventDate), timeutil.FormatDateID(v.ProductionDate), v.MenuCount, v.TotalPortions, v.Locked})
	}
	return out
}
func OrderDetails(v *model.OrderDetail) OrderDetail {
	return OrderDetail{v.ID, v.CustomerName, timeutil.FormatDateID(v.EventDate), timeutil.FormatDateID(v.ProductionDate), v.TotalPortions, v.Locked}
}
