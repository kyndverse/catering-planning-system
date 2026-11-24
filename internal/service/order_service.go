package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/timeutil"
	"golang-web/internal/validator"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderItemNotFound = errors.New("order item not found")
	ErrOrderLocked       = errors.New("order locked")
)

const lockedMessage = "Pesanan terkunci karena tanggal acara kurang dari 2 hari lagi"

type OrdersService struct {
	Log             *logrus.Logger
	DB              *sql.DB
	Validator       *validator.Validator
	OrderRepository *repository.OrderRepository
}

func NewOrdersService(log *logrus.Logger, db *sql.DB, v *validator.Validator, repo *repository.OrderRepository) *OrdersService {
	return &OrdersService{Log: log, DB: db, Validator: v, OrderRepository: repo}
}
func OrderLocked(eventDate, today time.Time) bool { return timeutil.DaysUntil(eventDate, today) < 2 }
func fieldError(field, message string) validator.ValidatorErrors {
	return validator.ValidatorErrors{field: message}
}
func parsePortions(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	return n, err == nil && n >= 1 && n <= 100000
}
func parseOrderDates(eventRaw, productionRaw string) (time.Time, time.Time, validator.ValidatorErrors) {
	event, err := timeutil.ParseDateISO(eventRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fieldError("EventDate", "Format tanggal tidak valid")
	}
	production := event
	if strings.TrimSpace(productionRaw) != "" {
		production, err = timeutil.ParseDateISO(productionRaw)
		if err != nil {
			return time.Time{}, time.Time{}, fieldError("ProductionDate", "Format tanggal tidak valid")
		}
	}
	if production.After(event) {
		return time.Time{}, time.Time{}, fieldError("ProductionDate", "Tanggal produksi tidak boleh setelah tanggal acara")
	}
	return event, production, nil
}
func (s *OrdersService) Search(ctx context.Context, req *model.SearchOrderRequest, now time.Time) ([]model.OrderSummary, validator.ValidatorErrors, error) {
	var ed, pd *time.Time
	errs := validator.ValidatorErrors{}
	if req.EventDate != "" {
		d, e := timeutil.ParseDateISO(req.EventDate)
		if e != nil {
			errs["EventDate"] = "Format tanggal tidak valid"
		} else {
			ed = &d
		}
	}
	if req.ProductionDate != "" {
		d, e := timeutil.ParseDateISO(req.ProductionDate)
		if e != nil {
			errs["ProductionDate"] = "Format tanggal tidak valid"
		} else {
			pd = &d
		}
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, nil, e
	}
	defer tx.Rollback()
	rows, e := s.OrderRepository.ListSummaries(tx, ed, pd)
	if e != nil {
		return nil, nil, e
	}
	today := timeutil.Today(now)
	for i := range rows {
		rows[i].Locked = OrderLocked(rows[i].EventDate, today)
	}
	if e = tx.Commit(); e != nil {
		return nil, nil, e
	}
	if len(errs) == 0 {
		return rows, nil, nil
	}
	return rows, errs, nil
}
func (s *OrdersService) GetOrderByID(ctx context.Context, id int64, now time.Time) (*model.OrderDetail, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	v, e := s.OrderRepository.FindOne(tx, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if e != nil {
		return nil, e
	}
	v.Locked = OrderLocked(v.EventDate, timeutil.Today(now))
	if e = tx.Commit(); e != nil {
		return nil, e
	}
	return v, nil
}
func (s *OrdersService) AddOrder(ctx context.Context, req *model.CreateOrderRequest, now time.Time) (int64, validator.ValidatorErrors, error) {
	if e := s.Validator.ValidateStruct(req); e != nil {
		return 0, e, nil
	}
	event, production, errs := parseOrderDates(req.EventDate, req.ProductionDate)
	if errs != nil {
		return 0, errs, nil
	}
	if timeutil.DaysUntil(event, timeutil.Today(now)) < 2 {
		return 0, fieldError("EventDate", "Tanggal acara minimal 2 hari dari sekarang"), nil
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return 0, nil, e
	}
	defer tx.Rollback()
	id, e := s.OrderRepository.Create(tx, req.CustomerName, event, production)
	if e != nil {
		return 0, nil, e
	}
	if e = tx.Commit(); e != nil {
		return 0, nil, e
	}
	return id, nil, nil
}
func (s *OrdersService) EditOrder(ctx context.Context, req *model.UpdateOrderRequest, now time.Time) (validator.ValidatorErrors, error) {
	if e := s.Validator.ValidateStruct(req); e != nil {
		return e, nil
	}
	today := timeutil.Today(now)
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	old, e := s.OrderRepository.FindOne(tx, req.ID)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if e != nil {
		return nil, e
	}
	if OrderLocked(old.EventDate, today) {
		return nil, ErrOrderLocked
	}
	event, production, errs := parseOrderDates(req.EventDate, req.ProductionDate)
	if errs != nil {
		return errs, nil
	}
	if timeutil.DaysUntil(event, today) < 2 {
		return fieldError("EventDate", "Tanggal acara minimal 2 hari dari sekarang"), nil
	}
	ok, e := s.OrderRepository.Update(tx, req.ID, req.CustomerName, event, production)
	if e != nil {
		return nil, e
	}
	if !ok {
		return nil, ErrOrderNotFound
	}
	return nil, tx.Commit()
}
func (s *OrdersService) DeleteOrder(ctx context.Context, req *model.DeleteOrderRequest, now time.Time) error {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	old, e := s.OrderRepository.FindOne(tx, req.ID)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrOrderNotFound
	}
	if e != nil {
		return e
	}
	if OrderLocked(old.EventDate, timeutil.Today(now)) {
		return ErrOrderLocked
	}
	ok, e := s.OrderRepository.Delete(tx, req.ID)
	if e != nil {
		return e
	}
	if !ok {
		return ErrOrderNotFound
	}
	return tx.Commit()
}
func (s *OrdersService) ListItemsByOrderID(ctx context.Context, id int64) ([]model.OrderItemView, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	ok, e := s.OrderRepository.OrderExists(tx, id)
	if e != nil {
		return nil, e
	}
	if !ok {
		return nil, ErrOrderNotFound
	}
	v, e := s.OrderRepository.ListItemsByOrderID(tx, id)
	if e != nil {
		return nil, e
	}
	return v, tx.Commit()
}
func (s *OrdersService) ListMenuOptions(ctx context.Context, id int64) ([]model.MenuOption, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	ok, e := s.OrderRepository.OrderExists(tx, id)
	if e != nil {
		return nil, e
	}
	if !ok {
		return nil, ErrOrderNotFound
	}
	v, e := s.OrderRepository.ListMenuOptions(tx, id)
	if e != nil {
		return nil, e
	}
	return v, tx.Commit()
}
func (s *OrdersService) AddOrderItem(ctx context.Context, req *model.AddOrderItemRequest, now time.Time) (validator.ValidatorErrors, error) {
	if e := s.Validator.ValidateStruct(req); e != nil {
		return e, nil
	}
	menuID, e := strconv.ParseInt(strings.TrimSpace(req.MenuID), 10, 64)
	if e != nil || menuID <= 0 {
		return fieldError("MenuID", "Menu tidak ditemukan"), nil
	}
	portions, ok := parsePortions(req.Portions)
	if !ok {
		return fieldError("Portions", "Jumlah porsi harus bilangan bulat antara 1 dan 100.000"), nil
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	order, e := s.OrderRepository.FindOne(tx, req.OrderID)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if e != nil {
		return nil, e
	}
	if OrderLocked(order.EventDate, timeutil.Today(now)) {
		return nil, ErrOrderLocked
	}
	kind, e := s.OrderRepository.FindRecipeType(tx, menuID)
	if errors.Is(e, sql.ErrNoRows) {
		return fieldError("MenuID", "Menu tidak ditemukan"), nil
	}
	if e != nil {
		return nil, e
	}
	if kind != "MENU" {
		return fieldError("MenuID", "Hanya resep bertipe Menu yang dapat dipesan"), nil
	}
	exists, e := s.OrderRepository.ItemExists(tx, req.OrderID, menuID)
	if e != nil {
		return nil, e
	}
	if exists {
		return fieldError("MenuID", "Menu sudah ada di pesanan ini. Ubah jumlah porsinya di tabel."), nil
	}
	if e = s.OrderRepository.AddItem(tx, req.OrderID, menuID, portions); e != nil {
		return nil, e
	}
	return nil, tx.Commit()
}
func (s *OrdersService) UpdateOrderItem(ctx context.Context, req *model.UpdateOrderItemRequest, now time.Time) (validator.ValidatorErrors, error) {
	if e := s.Validator.ValidateStruct(req); e != nil {
		return e, nil
	}
	portions, ok := parsePortions(req.Portions)
	if !ok {
		return fieldError("Portions", "Jumlah porsi harus bilangan bulat antara 1 dan 100.000"), nil
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	order, e := s.OrderRepository.FindOne(tx, req.OrderID)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if e != nil {
		return nil, e
	}
	if OrderLocked(order.EventDate, timeutil.Today(now)) {
		return nil, ErrOrderLocked
	}
	updated, e := s.OrderRepository.UpdateItemPortions(tx, req.OrderID, req.MenuID, portions)
	if e != nil {
		return nil, e
	}
	if !updated {
		return nil, ErrOrderItemNotFound
	}
	return nil, tx.Commit()
}
func (s *OrdersService) DeleteOrderItem(ctx context.Context, req *model.DeleteOrderItemRequest, now time.Time) error {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	order, e := s.OrderRepository.FindOne(tx, req.OrderID)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrOrderNotFound
	}
	if e != nil {
		return e
	}
	if OrderLocked(order.EventDate, timeutil.Today(now)) {
		return ErrOrderLocked
	}
	deleted, e := s.OrderRepository.DeleteItem(tx, req.OrderID, req.MenuID)
	if e != nil {
		return e
	}
	if !deleted {
		return ErrOrderItemNotFound
	}
	return tx.Commit()
}
