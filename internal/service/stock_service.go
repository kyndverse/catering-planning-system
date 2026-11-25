package service

import (
	"context"
	"database/sql"
	"errors"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/timeutil"
	"golang-web/internal/validator"
	"golang-web/internal/view"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const stockPageSize = 50

var stockQuantityPattern = regexp.MustCompile(`^\d{1,12}(\.\d{1,6})?$`)
var moneyPattern = regexp.MustCompile(`^\d{1,13}(\.\d{1,2})?$`)

type StockService struct {
	DB              *sql.DB
	Validator       *validator.Validator
	StockRepository *repository.StockRepository
}

func NewStockService(db *sql.DB, v *validator.Validator, r *repository.StockRepository) *StockService {
	return &StockService{db, v, r}
}
func ParsePositiveQuantity(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if !stockQuantityPattern.MatchString(s) {
		return "", false
	}
	v, ok := new(big.Rat).SetString(s)
	return s, ok && v.Sign() > 0
}
func ParseMoney(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if !moneyPattern.MatchString(s) {
		return "", false
	}
	v, ok := new(big.Rat).SetString(s)
	return s, ok && v.Sign() >= 0
}
func IsZeroMoney(s string) bool { v, ok := new(big.Rat).SetString(s); return ok && v.Sign() == 0 }
func SignedAdjustment(d, q string) (string, bool) {
	if d == "increase" {
		return q, true
	}
	if d == "decrease" {
		return "-" + q, true
	}
	return "", false
}
func NegateForUsage(q string) string { return "-" + q }
func (s *StockService) ListStocks(ctx context.Context, req model.SearchStockRequest) ([]model.StockRecord, int, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, 0, e
	}
	defer tx.Rollback()
	rows, e := s.StockRepository.ListStocks(tx, strings.TrimSpace(req.Q))
	if e != nil {
		return nil, 0, e
	}
	n := 0
	for _, x := range rows {
		if view.IsNegative(x.StockQty) {
			n++
		}
	}
	return rows, n, tx.Commit()
}
func (s *StockService) ListIngredientOptions(ctx context.Context) ([]model.IngredientOption, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	x, e := s.StockRepository.ListIngredientOptions(tx)
	if e != nil {
		return nil, e
	}
	return x, tx.Commit()
}
func filterError() validator.ValidatorErrors {
	return fieldError("Error", "Filter tidak valid dan diabaikan")
}
func (s *StockService) ListMovements(ctx context.Context, req model.SearchMovementRequest, now time.Time) ([]model.MovementRecord, int, int, validator.ValidatorErrors, error) {
	f := repository.MovementFilter{}
	var ferr validator.ValidatorErrors
	invalid := false
	if req.IngredientID != "" {
		if id, e := strconv.ParseInt(req.IngredientID, 10, 64); e == nil && id > 0 {
			f.IngredientID = req.IngredientID
		} else {
			invalid = true
		}
	}
	if req.Type != "" {
		if req.Type == "PURCHASE" || req.Type == "USAGE" || req.Type == "ADJUSTMENT" {
			f.Type = req.Type
		} else {
			invalid = true
		}
	}
	for _, p := range []struct {
		raw string
		dst *string
	}{{req.From, &f.From}, {req.To, &f.To}} {
		if p.raw != "" {
			if d, e := timeutil.ParseDateISO(p.raw); e == nil {
				*p.dst = timeutil.FormatDateISO(d)
			} else {
				invalid = true
			}
		}
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		invalid = true
		f.From = ""
		f.To = ""
	}
	page := 1
	if req.Page != "" {
		if n, e := strconv.Atoi(req.Page); e == nil && n > 0 {
			page = n
		} else {
			invalid = true
		}
	}
	if invalid {
		ferr = filterError()
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, 0, 0, ferr, e
	}
	defer tx.Rollback()
	total, e := s.StockRepository.CountMovements(tx, f)
	if e != nil {
		return nil, 0, 0, ferr, e
	}
	pages := (total + stockPageSize - 1) / stockPageSize
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}
	rows, e := s.StockRepository.ListMovements(tx, f, stockPageSize, (page-1)*stockPageSize)
	if e != nil {
		return nil, 0, 0, ferr, e
	}
	return rows, page, pages, ferr, tx.Commit()
}
func parseStockDate(raw string, now time.Time) (time.Time, validator.ValidatorErrors) {
	d, e := timeutil.ParseDateISO(raw)
	if e != nil {
		return time.Time{}, fieldError("MovementDate", "Format tanggal tidak valid")
	}
	if d.After(timeutil.Today(now)) {
		return time.Time{}, fieldError("MovementDate", "Tanggal tidak boleh setelah hari ini")
	}
	return d, nil
}
func parseStockID(raw string) (int64, validator.ValidatorErrors) {
	id, e := strconv.ParseInt(raw, 10, 64)
	if e != nil || id <= 0 {
		return 0, fieldError("IngredientID", "Bahan baku tidak ditemukan")
	}
	return id, nil
}
func merge(a, b validator.ValidatorErrors) validator.ValidatorErrors {
	if a == nil {
		return b
	}
	for k, v := range b {
		a[k] = v
	}
	return a
}
func (s *StockService) validateBase(date, id, qty string, now time.Time) (time.Time, int64, string, validator.ValidatorErrors) {
	d, e := parseStockDate(date, now)
	x, ie := parseStockID(id)
	q, ok := ParsePositiveQuantity(qty)
	if !ok {
		ie = merge(ie, fieldError("Quantity", "Jumlah harus angka lebih dari 0 (maksimal 6 angka desimal)"))
	}
	return d, x, q, merge(e, ie)
}
func (s *StockService) RecordPurchase(ctx context.Context, r model.CreatePurchaseRequest, now time.Time) (*model.StockWriteResult, validator.ValidatorErrors, error) {
	d, id, q, fe := s.validateBase(r.MovementDate, r.IngredientID, r.Quantity, now)
	cost, ok := ParseMoney(r.TotalCost)
	if !ok {
		fe = merge(fe, fieldError("TotalCost", "Total harga harus angka 0 atau lebih (maksimal 2 angka desimal)"))
	}
	note := strings.TrimSpace(r.Note)
	if len(note) > 200 {
		fe = merge(fe, fieldError("Note", "Catatan maksimal 200 karakter"))
	}
	if r.UpdatePrice == "1" && ok && IsZeroMoney(cost) {
		fe = merge(fe, fieldError("TotalCost", "Total harga harus lebih dari 0 untuk memperbarui harga acuan"))
	}
	if fe != nil {
		return nil, fe, nil
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, nil, e
	}
	defer tx.Rollback()
	_, e = s.StockRepository.LockIngredient(tx, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, fieldError("IngredientID", "Bahan baku tidak ditemukan"), nil
	}
	if e != nil {
		return nil, nil, e
	}
	if r.UpdatePrice == "1" {
		exists, e := s.StockRepository.PriceExistsOnOrAfter(tx, id, d)
		if e != nil {
			return nil, nil, e
		}
		if exists {
			return nil, fieldError("UpdatePrice", "Sudah ada harga acuan pada atau setelah tanggal ini. Hapus centang \"Perbarui harga acuan\" atau pakai tanggal yang lebih baru."), nil
		}
	}
	if e = s.StockRepository.InsertPurchase(tx, id, d, q, cost, note); e != nil {
		return nil, nil, e
	}
	if r.UpdatePrice == "1" {
		if e = s.StockRepository.InsertPrice(tx, id, cost, q, d); e != nil {
			return nil, nil, e
		}
	}
	if e = tx.Commit(); e != nil {
		return nil, nil, e
	}
	return &model.StockWriteResult{IngredientID: id}, nil, nil
}
func (s *StockService) RecordUsage(ctx context.Context, r model.CreateUsageRequest, now time.Time) (*model.StockWriteResult, validator.ValidatorErrors, error) {
	d, id, q, fe := s.validateBase(r.MovementDate, r.IngredientID, r.Quantity, now)
	p, pe := timeutil.ParseDateISO(r.ProductionDate)
	if pe != nil {
		fe = merge(fe, fieldError("ProductionDate", "Format tanggal tidak valid"))
	}
	note := strings.TrimSpace(r.Note)
	if len(note) > 200 {
		fe = merge(fe, fieldError("Note", "Catatan maksimal 200 karakter"))
	}
	if fe != nil {
		return nil, fe, nil
	}
	return s.recordNegative(ctx, id, d, NegateForUsage(q), p, note, r.ConfirmNegative, now, "USAGE")
}
func (s *StockService) RecordAdjustment(ctx context.Context, r model.CreateAdjustmentRequest, now time.Time) (*model.StockWriteResult, validator.ValidatorErrors, error) {
	d, id, q, fe := s.validateBase(r.MovementDate, r.IngredientID, r.Quantity, now)
	signed, ok := SignedAdjustment(r.Direction, q)
	if !ok {
		fe = merge(fe, fieldError("Direction", "Pilih jenis koreksi"))
	}
	note := strings.TrimSpace(r.Note)
	if note == "" {
		fe = merge(fe, fieldError("Note", "Alasan koreksi wajib diisi"))
	}
	if len(note) > 500 {
		fe = merge(fe, fieldError("Note", "Catatan maksimal 500 karakter"))
	}
	if fe != nil {
		return nil, fe, nil
	}
	return s.recordNegative(ctx, id, d, signed, time.Time{}, note, r.ConfirmNegative, now, "ADJUSTMENT")
}
func (s *StockService) recordNegative(ctx context.Context, id int64, d time.Time, delta string, p time.Time, note, confirm string, now time.Time, typ string) (*model.StockWriteResult, validator.ValidatorErrors, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, nil, e
	}
	defer tx.Rollback()
	ing, e := s.StockRepository.LockIngredient(tx, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, fieldError("IngredientID", "Bahan baku tidak ditemukan"), nil
	}
	if e != nil {
		return nil, nil, e
	}
	projected, negative, e := s.StockRepository.ProjectedStock(tx, id, delta)
	if e != nil {
		return nil, nil, e
	}
	if negative && confirm != "1" {
		return &model.StockWriteResult{IngredientID: id, Warning: "Stok " + ing.Name + " akan menjadi " + view.FormatQuantity(projected) + " " + ing.BaseUnit + ". Pilih \"Tetap Catat\" jika memang sudah sesuai."}, nil, nil
	}
	if typ == "USAGE" {
		e = s.StockRepository.InsertUsage(tx, id, d, delta, p, note)
	} else {
		e = s.StockRepository.InsertAdjustment(tx, id, d, delta, note)
	}
	if e != nil {
		return nil, nil, e
	}
	if e = tx.Commit(); e != nil {
		return nil, nil, e
	}
	return &model.StockWriteResult{IngredientID: id}, nil, nil
}
