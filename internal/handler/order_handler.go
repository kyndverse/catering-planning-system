package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/internal/timeutil"
	"golang-web/internal/validator"
	orderView "golang-web/internal/view"
	"golang-web/web"
)

type OrderHandler struct {
	Log          *logrus.Logger
	Renderer     *web.Renderer
	Decoder      *form.Decoder
	OrderService *service.OrdersService
}

const orderLockedMessage = "Pesanan terkunci karena tanggal acara kurang dari 2 hari lagi"

func NewOrderHandler(log *logrus.Logger, r *web.Renderer, d *form.Decoder, s *service.OrdersService) *OrderHandler {
	return &OrderHandler{log, r, d, s}
}
func (h *OrderHandler) id(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	id, e := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	if e != nil || id <= 0 {
		http.Error(w, "Parameter tidak valid", 400)
		return 0, false
	}
	return id, true
}
func (h *OrderHandler) Search(w http.ResponseWriter, r *http.Request) {
	req := &model.SearchOrderRequest{EventDate: r.URL.Query().Get("event_date"), ProductionDate: r.URL.Query().Get("production_date")}
	rows, errs, e := h.OrderService.Search(r.Context(), req, time.Now())
	if e != nil {
		h.fail(w, e, "gagal mengambil pesanan")
		return
	}
	h.Renderer.Render(w, "orders/index.html", map[string]any{"title": "Pesanan", "active": "orders", "filter": req, "orders": orderView.OrderSummaries(rows), "errors": errs})
}
func (h *OrderHandler) NewOrderPage(w http.ResponseWriter, r *http.Request) {
	h.renderOrderForm(w, false, 0, model.CreateOrderRequest{}, nil, time.Now())
}
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	req := new(model.CreateOrderRequest)
	if !h.decode(w, r, req) {
		return
	}
	id, errs, e := h.OrderService.AddOrder(r.Context(), req, time.Now())
	if e != nil {
		h.fail(w, e, "gagal membuat pesanan")
		return
	}
	if errs != nil {
		h.renderOrderForm(w, false, 0, *req, errs, time.Now())
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%d", id), http.StatusSeeOther)
}
func (h *OrderHandler) GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "id")
	if ok {
		h.renderOrderDetail(w, r, id, newOrderItemForm(id), nil)
	}
}
func (h *OrderHandler) EditOrderPage(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "id")
	if !ok {
		return
	}
	order, e := h.OrderService.GetOrderByID(r.Context(), id, time.Now())
	if errors.Is(e, service.ErrOrderNotFound) {
		http.NotFound(w, r)
		return
	}
	if e != nil {
		h.fail(w, e, "gagal mengambil pesanan")
		return
	}
	if order.Locked {
		http.Redirect(w, r, fmt.Sprintf("/orders/%d", id), http.StatusSeeOther)
		return
	}
	h.renderOrderForm(w, true, id, model.CreateOrderRequest{CustomerName: order.CustomerName, EventDate: timeutil.FormatDateISO(order.EventDate), ProductionDate: timeutil.FormatDateISO(order.ProductionDate)}, nil, time.Now())
}
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "id")
	if !ok {
		return
	}
	req := new(model.UpdateOrderRequest)
	if !h.decode(w, r, req) {
		return
	}
	req.ID = id
	errs, e := h.OrderService.EditOrder(r.Context(), req, time.Now())
	if errors.Is(e, service.ErrOrderNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(e, service.ErrOrderLocked) {
		h.renderOrderDetail(w, r, id, newOrderItemForm(id), validator.ValidatorErrors{"Error": orderLockedMessage})
		return
	}
	if e != nil {
		h.fail(w, e, "gagal mengubah pesanan")
		return
	}
	if errs != nil {
		h.renderOrderForm(w, true, id, model.CreateOrderRequest{CustomerName: req.CustomerName, EventDate: req.EventDate, ProductionDate: req.ProductionDate}, errs, time.Now())
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%d", id), http.StatusSeeOther)
}
func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "id")
	if !ok {
		return
	}
	e := h.OrderService.DeleteOrder(r.Context(), &model.DeleteOrderRequest{ID: id}, time.Now())
	if errors.Is(e, service.ErrOrderNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(e, service.ErrOrderLocked) {
		rows, errs, se := h.OrderService.Search(r.Context(), &model.SearchOrderRequest{}, time.Now())
		if se != nil {
			h.fail(w, se, "gagal mengambil pesanan")
			return
		}
		if errs == nil {
			errs = validator.ValidatorErrors{}
		}
		errs["Error"] = orderLockedMessage
		h.Renderer.Render(w, "orders/index.html", map[string]any{"title": "Pesanan", "active": "orders", "filter": model.SearchOrderRequest{}, "orders": orderView.OrderSummaries(rows), "errors": errs})
		return
	}
	if e != nil {
		h.fail(w, e, "gagal menghapus pesanan")
		return
	}
	http.Redirect(w, r, "/orders", http.StatusSeeOther)
}
func (h *OrderHandler) AddOrderItem(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "id")
	if !ok {
		return
	}
	req := new(model.AddOrderItemRequest)
	if !h.decode(w, r, req) {
		return
	}
	req.OrderID = id
	errs, e := h.OrderService.AddOrderItem(r.Context(), req, time.Now())
	itemForm := newOrderItemForm(id)
	itemForm.Data = *req
	if h.itemError(w, r, id, e, errs, itemForm) {
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%d", id), http.StatusSeeOther)
}
func (h *OrderHandler) UpdateOrderItem(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "id")
	if !ok {
		return
	}
	menuID, ok := h.id(w, r, "menuID")
	if !ok {
		return
	}
	req := new(model.UpdateOrderItemRequest)
	if !h.decode(w, r, req) {
		return
	}
	req.OrderID = id
	req.MenuID = menuID
	errs, e := h.OrderService.UpdateOrderItem(r.Context(), req, time.Now())
	if h.itemError(w, r, id, e, errs, newOrderItemForm(id)) {
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%d", id), http.StatusSeeOther)
}
func (h *OrderHandler) DeleteOrderItem(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r, "id")
	if !ok {
		return
	}
	menuID, ok := h.id(w, r, "menuID")
	if !ok {
		return
	}
	e := h.OrderService.DeleteOrderItem(r.Context(), &model.DeleteOrderItemRequest{OrderID: id, MenuID: menuID}, time.Now())
	if errors.Is(e, service.ErrOrderNotFound) || errors.Is(e, service.ErrOrderItemNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(e, service.ErrOrderLocked) {
		h.renderOrderDetail(w, r, id, newOrderItemForm(id), validator.ValidatorErrors{"Error": orderLockedMessage})
		return
	}
	if e != nil {
		h.fail(w, e, "gagal menghapus menu pesanan")
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%d", id), http.StatusSeeOther)
}

func (h *OrderHandler) itemError(w http.ResponseWriter, r *http.Request, id int64, e error, errs validator.ValidatorErrors, itemForm model.FormData[model.AddOrderItemRequest]) bool {
	if errors.Is(e, service.ErrOrderNotFound) || errors.Is(e, service.ErrOrderItemNotFound) {
		http.NotFound(w, r)
		return true
	}
	if errors.Is(e, service.ErrOrderLocked) {
		h.renderOrderDetail(w, r, id, itemForm, validator.ValidatorErrors{"Error": orderLockedMessage})
		return true
	}
	if e != nil {
		h.fail(w, e, "gagal mengubah menu pesanan")
		return true
	}
	if errs != nil {
		h.renderOrderDetail(w, r, id, itemForm, errs)
		return true
	}
	return false
}
func (h *OrderHandler) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if e := r.ParseForm(); e != nil {
		http.Error(w, "Data formulir tidak valid", 400)
		return false
	}
	if e := h.Decoder.Decode(dst, r.PostForm); e != nil {
		http.Error(w, "Data formulir tidak valid", 400)
		return false
	}
	return true
}
func (h *OrderHandler) renderOrderForm(w http.ResponseWriter, edit bool, id int64, data model.CreateOrderRequest, errs validator.ValidatorErrors, now time.Time) {
	title, label, cancel, action := "Tambah Pesanan", "Simpan", "/orders", "/orders"
	if edit {
		title = "Ubah Pesanan"
		label = "Simpan Perubahan"
		cancel = fmt.Sprintf("/orders/%d", id)
		action = fmt.Sprintf("/orders/%d", id)
	}
	form := model.FormData[model.CreateOrderRequest]{Action: action, Data: data, Errors: errs}
	h.Renderer.Render(w, "orders/form.html", map[string]any{"title": title, "active": "orders", "submitLabel": label, "cancelURL": cancel, "minEventDate": timeutil.FormatDateISO(timeutil.Today(now).AddDate(0, 0, 2)), "form": form})
}
func newOrderItemForm(id int64) model.FormData[model.AddOrderItemRequest] {
	return model.FormData[model.AddOrderItemRequest]{Action: fmt.Sprintf("/orders/%d/items", id)}
}
func (h *OrderHandler) renderOrderDetail(w http.ResponseWriter, r *http.Request, id int64, itemForm model.FormData[model.AddOrderItemRequest], pageErrors validator.ValidatorErrors) {
	order, e := h.OrderService.GetOrderByID(r.Context(), id, time.Now())
	if errors.Is(e, service.ErrOrderNotFound) {
		http.NotFound(w, r)
		return
	}
	if e != nil {
		h.fail(w, e, "gagal mengambil detail pesanan")
		return
	}
	items, e := h.OrderService.ListItemsByOrderID(r.Context(), id)
	if e != nil {
		h.fail(w, e, "gagal mengambil menu pesanan")
		return
	}
	options, e := h.OrderService.ListMenuOptions(r.Context(), id)
	if e != nil {
		h.fail(w, e, "gagal mengambil pilihan menu")
		return
	}
	h.Renderer.Render(w, "orders/show.html", map[string]any{"title": "Detail Pesanan", "active": "orders", "order": orderView.OrderDetails(order), "items": items, "menuOptions": options, "itemForm": itemForm, "errors": pageErrors})
}
func (h *OrderHandler) fail(w http.ResponseWriter, e error, msg string) {
	h.Log.WithError(e).Error(msg)
	http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
}
