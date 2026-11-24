package view

import (
	"golang-web/internal/model"
	"testing"
	"time"
)

func TestOrderViewsFormatDatesAndKeepLock(t *testing.T) {
	d := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	summaries := OrderSummaries([]model.OrderSummary{{ID: 7, CustomerName: "Klien", EventDate: d, ProductionDate: d, MenuCount: 1, TotalPortions: 10, Locked: true}})
	if len(summaries) != 1 || summaries[0].EventDate != "1 Des 2026" || !summaries[0].Locked {
		t.Fatalf("unexpected summary: %+v", summaries)
	}
	detail := OrderDetails(&model.OrderDetail{ID: 7, EventDate: d, ProductionDate: d, Locked: true})
	if detail.EventDate != "1 Des 2026" || !detail.Locked {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}
