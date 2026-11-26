package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/timeutil"
	"math/big"
	"time"
)

var UsageVarianceTolerance = big.NewRat(5, 100)

type UsageReportService struct {
	Log                   *logrus.Logger
	DB                    *sql.DB
	UsageReportRepository *repository.UsageReportRepository
	OrderRepository       *repository.OrderRepository
}

func NewUsageReportService(log *logrus.Logger, db *sql.DB, repo *repository.UsageReportRepository, orders *repository.OrderRepository) *UsageReportService {
	return &UsageReportService{log, db, repo, orders}
}
func (s *UsageReportService) ListUsageDates(ctx context.Context, now time.Time) ([]model.UsageDateSummary, error) {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	x, err := s.UsageReportRepository.ListUsageDates(tx, timeutil.Today(now))
	if err != nil {
		return nil, err
	}
	return x, tx.Commit()
}
func (s *UsageReportService) GetUsageReport(ctx context.Context, date, now time.Time) (*model.UsageReport, error) {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	orders, err := s.OrderRepository.ListSummaries(tx, nil, &date)
	if err != nil {
		return nil, err
	}
	total := 0
	for _, o := range orders {
		total += o.TotalPortions
	}
	rows, err := s.UsageReportRepository.ListPlanVsActual(tx, date, timeutil.Today(now))
	if err != nil {
		return nil, err
	}
	lines, summary, err := BuildUsage(rows, UsageVarianceTolerance)
	if err != nil {
		return nil, err
	}
	out := &model.UsageReport{Date: date, OrderCount: len(orders), TotalPortions: total, Lines: lines, Summary: summary}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}
func BuildUsage(rows []model.UsageInputRow, tolerance *big.Rat) ([]model.UsageLine, model.UsageSummary, error) {
	if tolerance == nil || tolerance.Sign() < 0 {
		return nil, model.UsageSummary{}, fmt.Errorf("invalid tolerance")
	}
	out := make([]model.UsageLine, 0, len(rows))
	summary := model.UsageSummary{TotalValue: new(big.Rat)}
	for _, row := range rows {
		planned, err := parseRat(row.PlannedQty)
		if err != nil {
			return nil, summary, err
		}
		actual, err := parseRat(row.ActualQty)
		if err != nil {
			return nil, summary, err
		}
		line := model.UsageLine{IngredientName: row.IngredientName, Unit: row.BaseUnit, Planned: planned, Actual: actual}
		if !row.HasActual && planned.Sign() > 0 {
			line.Status = "unrecorded"
			summary.UnrecordedCount++
			out = append(out, line)
			continue
		}
		if planned.Sign() == 0 {
			line.Status = "unplanned"
		} else {
			line.Variance = new(big.Rat).Sub(actual, planned)
			line.Pct = new(big.Rat).Quo(line.Variance, planned)
			abs := new(big.Rat).Abs(line.Pct)
			if abs.Cmp(tolerance) <= 0 {
				line.Status = "match"
			} else if line.Pct.Sign() > 0 {
				line.Status = "over"
				summary.OverCount++
			} else {
				line.Status = "under"
				summary.UnderCount++
			}
		}
		if line.Variance == nil {
			line.Variance = new(big.Rat).Sub(actual, planned)
		}
		if row.PricePerPurchaseUnit == nil {
			summary.PriceIncomplete = true
		} else {
			price, err := parseRat(*row.PricePerPurchaseUnit)
			if err != nil {
				return nil, summary, err
			}
			base, err := parseRat(row.BaseQtyPerPurchaseUnit)
			if err != nil || base.Sign() <= 0 {
				return nil, summary, fmt.Errorf("invalid base_qty_per_purchase_unit %q", row.BaseQtyPerPurchaseUnit)
			}
			line.Value = new(big.Rat).Quo(new(big.Rat).Mul(line.Variance, price), base)
			summary.TotalValue.Add(summary.TotalValue, line.Value)
			summary.HasValue = true
		}
		out = append(out, line)
	}
	return out, summary, nil
}
