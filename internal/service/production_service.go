package service

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/timeutil"
)

type ProductionService struct {
	Log                  *logrus.Logger
	DB                   *sql.DB
	ProductionRepository *repository.ProductionRepository
	OrderRepository      *repository.OrderRepository
}

func NewProductionService(log *logrus.Logger, db *sql.DB, production *repository.ProductionRepository, orders *repository.OrderRepository) *ProductionService {
	return &ProductionService{log, db, production, orders}
}

func (s *ProductionService) ListUpcoming(ctx context.Context, now time.Time) ([]model.ProductionDateSummary, error) {
	today := timeutil.Today(now)
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := s.ProductionRepository.ListUpcomingDates(tx, today)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].AllLocked = OrderLocked(rows[i].MaxEventDate, today)
	}
	return rows, tx.Commit()
}

func (s *ProductionService) GetPlan(ctx context.Context, date, now time.Time) (*model.ProductionPlan, error) {
	today := timeutil.Today(now)
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	orders, err := s.OrderRepository.ListSummaries(tx, nil, &date)
	if err != nil {
		return nil, err
	}
	total, editable := 0, false
	for i := range orders {
		orders[i].Locked = OrderLocked(orders[i].EventDate, today)
		total += orders[i].TotalPortions
		editable = editable || !orders[i].Locked
	}
	menus, err := s.ProductionRepository.ListMenuTotals(tx, date)
	if err != nil {
		return nil, err
	}
	rows, err := s.ProductionRepository.ListRequirements(tx, date, today)
	if err != nil {
		return nil, err
	}
	shopping, err := BuildShoppingList(rows)
	if err != nil {
		return nil, err
	}
	plan := &model.ProductionPlan{Date: date, Orders: orders, Menus: menus, TotalPortions: total, HasEditableOrders: editable, Requirements: BuildRequirements(rows), Shopping: shopping}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return plan, nil
}

func parseRat(raw string) (*big.Rat, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(raw))
	if !ok {
		return nil, fmt.Errorf("invalid decimal %q", raw)
	}
	return r, nil
}
func ratString(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	d := new(big.Int).Set(r.Denom())
	twos, fives := 0, 0
	divisor := big.NewInt(2)
	rem := new(big.Int)
	for {
		rem.Mod(d, divisor)
		if rem.Sign() != 0 {
			break
		}
		d.Quo(d, divisor)
		twos++
	}
	divisor.SetInt64(5)
	for {
		rem.Mod(d, divisor)
		if rem.Sign() != 0 {
			break
		}
		d.Quo(d, divisor)
		fives++
	}
	if d.Cmp(big.NewInt(1)) != 0 {
		return r.RatString()
	}
	scale := twos
	if fives > scale {
		scale = fives
	}
	n := new(big.Int).Set(r.Num())
	for i := 0; i < scale-twos; i++ {
		n.Mul(n, big.NewInt(2))
	}
	for i := 0; i < scale-fives; i++ {
		n.Mul(n, big.NewInt(5))
	}
	sign := ""
	if n.Sign() < 0 {
		sign = "-"
		n.Abs(n)
	}
	text := n.String()
	if len(text) <= scale {
		text = strings.Repeat("0", scale-len(text)+1) + text
	}
	out := sign + text[:len(text)-scale] + "." + text[len(text)-scale:]
	return strings.TrimRight(strings.TrimRight(out, "0"), ".")
}
func ceilDiv(a, b *big.Rat) *big.Int {
	q := new(big.Rat).Quo(a, b)
	quo, rem := new(big.Int).QuoRem(q.Num(), q.Denom(), new(big.Int))
	if rem.Sign() > 0 {
		quo.Add(quo, big.NewInt(1))
	}
	return quo
}

func BuildRequirements(rows []model.ProductionRequirementRow) []model.ProductionRequirement {
	out := make([]model.ProductionRequirement, 0, len(rows))
	for _, row := range rows {
		required, _ := parseRat(row.RequiredQty)
		stock, _ := parseRat(row.StockQty)
		short := false
		if required != nil && stock != nil {
			short = required.Cmp(stock) > 0
		}
		out = append(out, model.ProductionRequirement{Name: row.Name, Unit: row.BaseUnit, Required: row.RequiredQty, Stock: row.StockQty, Short: short})
	}
	return out
}
func BuildShoppingList(rows []model.ProductionRequirementRow) (model.ShoppingList, error) {
	out := model.ShoppingList{Items: make([]model.ShoppingItem, 0)}
	total := new(big.Rat)
	for _, row := range rows {
		required, err := parseRat(row.RequiredQty)
		if err != nil {
			return model.ShoppingList{}, err
		}
		stock, err := parseRat(row.StockQty)
		if err != nil {
			return model.ShoppingList{}, err
		}
		base, err := parseRat(row.BaseQtyPerPurchaseUnit)
		if err != nil || base.Sign() <= 0 {
			return model.ShoppingList{}, fmt.Errorf("invalid base_qty_per_purchase_unit %q", row.BaseQtyPerPurchaseUnit)
		}
		shortage := new(big.Rat).Sub(required, stock)
		if shortage.Sign() <= 0 {
			continue
		}
		buy := ceilDiv(shortage, base)
		if !buy.IsInt64() {
			return model.ShoppingList{}, fmt.Errorf("buy quantity overflows int64 for %q", row.Name)
		}
		item := model.ShoppingItem{Name: row.Name, Unit: row.BaseUnit, Shortage: ratString(shortage), BuyQty: buy.Int64(), PurchaseUnit: row.PurchaseUnit}
		if row.PricePerPurchaseUnit == nil {
			out.PriceIncomplete = true
		} else {
			price, err := parseRat(*row.PricePerPurchaseUnit)
			if err != nil {
				return model.ShoppingList{}, err
			}
			subtotal := new(big.Rat).Mul(new(big.Rat).SetInt(buy), price)
			item.HasPrice = true
			item.PricePerPurchaseUnit = ratString(price)
			item.Subtotal = ratString(subtotal)
			total.Add(total, subtotal)
		}
		out.Items = append(out.Items, item)
	}
	out.TotalCost = ratString(total)
	return out, nil
}
