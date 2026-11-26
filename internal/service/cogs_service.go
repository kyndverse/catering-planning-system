package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/timeutil"
	"math/big"
	"strconv"
	"time"
)

var ErrMenuNotFound = errors.New("menu not found")

type CogsService struct {
	Log            *logrus.Logger
	DB             *sql.DB
	CogsRepository *repository.CogsRepository
}

func NewCogsService(log *logrus.Logger, db *sql.DB, repo *repository.CogsRepository) *CogsService {
	return &CogsService{log, db, repo}
}
func (s *CogsService) ListMenuCogs(ctx context.Context, now time.Time) ([]model.MenuCogs, error) {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	menus, err := s.CogsRepository.ListMenus(tx)
	if err != nil {
		return nil, err
	}
	rows, err := s.CogsRepository.ListMenuIngredientCosts(tx, "", timeutil.Today(now))
	if err != nil {
		return nil, err
	}
	grouped := map[int64][]model.MenuIngredientCostRow{}
	for _, r := range rows {
		grouped[r.MenuID] = append(grouped[r.MenuID], r)
	}
	out := make([]model.MenuCogs, 0, len(menus))
	for _, m := range menus {
		x, e := BuildMenuCogs(m, grouped[m.ID])
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, tx.Commit()
}
func (s *CogsService) GetMenuCogs(ctx context.Context, id int64, now time.Time) (*model.MenuCogs, error) {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	m, err := s.CogsRepository.FindMenu(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMenuNotFound
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.CogsRepository.ListMenuIngredientCosts(tx, strconv.FormatInt(id, 10), timeutil.Today(now))
	if err != nil {
		return nil, err
	}
	x, err := BuildMenuCogs(*m, rows)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &x, nil
}
func BuildMenuCogs(menu model.MenuRef, rows []model.MenuIngredientCostRow) (model.MenuCogs, error) {
	out := model.MenuCogs{Menu: menu, Lines: make([]model.CostLine, 0, len(rows)), Total: new(big.Rat)}
	for _, row := range rows {
		qty, err := parseRat(row.QtyPerPortion)
		if err != nil {
			return out, err
		}
		base, err := parseRat(row.BaseQtyPerPurchaseUnit)
		if err != nil || base.Sign() <= 0 {
			return out, fmt.Errorf("invalid base_qty_per_purchase_unit %q", row.BaseQtyPerPurchaseUnit)
		}
		line := model.CostLine{IngredientName: row.IngredientName, Unit: row.BaseUnit, Qty: qty, PriceSince: row.PriceSince}
		if row.PricePerPurchaseUnit == nil {
			out.Missing++
		} else {
			price, err := parseRat(*row.PricePerPurchaseUnit)
			if err != nil {
				return out, err
			}
			line.PricePerBase = new(big.Rat).Quo(price, base)
			line.Cost = new(big.Rat).Quo(new(big.Rat).Mul(qty, price), base)
			out.Total.Add(out.Total, line.Cost)
		}
		out.Lines = append(out.Lines, line)
	}
	if len(rows) == 0 {
		out.Status = "empty"
	} else if out.Missing > 0 {
		out.Status = "incomplete"
	} else {
		out.Status = "complete"
	}
	return out, nil
}
