package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type InvoicesApiHandler struct {
	db             *gorm.DB
	config         *conf.Config
	invoiceService services.InvoiceService
	userService    services.IUserService
	summaryService services.ISummaryService
	clientService  services.IClientService
}

func NewInvoicesApiHandler(db *gorm.DB, invoiceService services.InvoiceService, userService services.IUserService, summaryService services.ISummaryService, clientService services.IClientService) *InvoicesApiHandler {
	return &InvoicesApiHandler{
		db:             db,
		invoiceService: invoiceService,
		userService:    userService,
		config:         conf.Get(),
		clientService:  clientService,
		summaryService: summaryService,
	}
}

func (h *InvoicesApiHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userService).Handler)
		r.Post("/compat/wakatime/v1/users/{user}/invoices", h.Create)
		r.Get("/compat/wakatime/v1/users/{user}/invoices", h.FetchUserInvoices)
		r.Get("/compat/wakatime/v1/users/{user}/invoices/{id}", h.GetInvoice)
		r.Put("/compat/wakatime/v1/users/{user}/invoices/{id}", h.UpdateInvoice)
		r.Delete("/compat/wakatime/v1/users/{user}/invoices/{id}", h.DeleteInvoice)
	})
}

func (h *InvoicesApiHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	InvoiceID := chi.URLParam(r, "id")

	if InvoiceID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	var params = &models.InvoiceUpdate{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid Input",
			"status":  http.StatusBadRequest,
		})
	}

	Invoice, err := h.invoiceService.GetInvoiceForUser(InvoiceID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invoice Cannot Be Found",
			"status":  http.StatusNotFound,
		})
		return
	}

	_, err = h.invoiceService.Update(Invoice, params)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "Error updating Invoice",
			"error_message": err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": Invoice,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (h *InvoicesApiHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	InvoiceID := chi.URLParam(r, "id")

	if InvoiceID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	Invoice, err := h.invoiceService.GetInvoiceForUser(InvoiceID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invoice Cannot Be Found",
			"status":  http.StatusBadRequest,
		})
		return
	}
	response := map[string]interface{}{
		"data": Invoice,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (h *InvoicesApiHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	invoiceID := chi.URLParam(r, "id")

	if invoiceID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	err := h.invoiceService.DeleteInvoice(invoiceID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invoice Cannot Be Deleted",
			"status":  http.StatusBadRequest,
		})
		return
	}
	response := map[string]interface{}{
		"message": "Invoice deleted successfully",
	}
	helpers.RespondJSON(w, r, http.StatusAccepted, response)
}

func (h *InvoicesApiHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)

	var params = &models.Invoice{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil || params.ClientID == "" || len(params.LineItems) == 0 {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid Input: client_id, line_items is required",
			"status":  http.StatusBadRequest,
		})
		return
	}

	client, err := h.clientService.GetClientForUser(params.ClientID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Client Cannot Be Found",
			"status":  http.StatusNotFound,
		})
		return
	}

	params.UserID = user.ID
	params.ID = uuid.NewV4().String()

	_, err = h.invoiceService.Create(params)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "An unexpected error occurred. Try again later",
			"error_message": err.Error(),
		})
		return
	}
	params.Client = *client
	response := map[string]interface{}{
		"data": params,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (h *InvoicesApiHandler) FetchUserInvoices(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	query := r.URL.Query().Get("q")

	Invoices, err := h.invoiceService.FetchUserInvoices(user.ID, query)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Error fetching Invoices. Try later.",
			"error":   err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": Invoices,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}
