package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/duke-git/lancet/v2/slice"
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
			"message":       "Invalid Input",
			"status":        http.StatusBadRequest,
			"error_message": err.Error(),
		})
	}

	invoice, err := h.invoiceService.GetInvoiceForUser(InvoiceID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invoice Cannot Be Found",
			"status":  http.StatusNotFound,
		})
		return
	}

	_, err = h.invoiceService.Update(invoice, params)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "Error updating Invoice",
			"error_message": err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": invoice,
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

	var params = &models.NewInvoiceData{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil || params.ClientID == "" || params.StartDate == "" || params.EndDate == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid Input: client_id, start_date, end_date is required",
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

	generatedData, err := h.FetchInvoiceLineItems(user, client, params.StartDate, params.EndDate)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": err.Error(),
			"status":  http.StatusNotFound,
		})
		return
	}

	newInvoice := &models.Invoice{
		ID:             uuid.NewV4().String(),
		UserID:         user.ID,
		ClientID:       params.ClientID,
		StartDate:      generatedData.StartDate,
		EndDate:        generatedData.EndDate,
		LineItems:      generatedData.LineItems,
		InvoiceSummary: fmt.Sprintf("Invoice for work done from %s to %s", helpers.FormatDate(generatedData.StartDate), helpers.FormatDate(generatedData.EndDate)),
		Destination:    client.Name,
		Origin:         fmt.Sprintf("%s \n Freelancer", user.Email),
		Tax:            0,
	}

	_, err = h.invoiceService.Create(newInvoice)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "An unexpected error occurred. Try again later",
			"error_message": err.Error(),
		})
		return
	}
	newInvoice.Client = *client
	helpers.RespondJSON(w, r, http.StatusCreated, map[string]interface{}{
		"message": "New Invoice Created",
		"status":  http.StatusCreated,
		"data":    &newInvoice,
	})
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

func (h *InvoicesApiHandler) FetchInvoiceLineItems(user *models.User, client *models.Client, start_date string, end_date string) (*models.NewlyGeneratedInvoiceData, error) {
	start, err := helpers.ParseDateTimeTZ(start_date, user.TZ())
	if err != nil {
		return nil, fmt.Errorf("invalid date %s provided", start_date)
	}

	end, err := helpers.ParseDateTimeTZ(end_date, user.TZ())
	if err != nil {
		return nil, fmt.Errorf("invalid date %s provided", end_date)
	}

	summary, err := h.clientService.FetchClientInvoiceLineItems(client, user, h.summaryService, start, end)
	if err != nil {
		return nil, fmt.Errorf("error fetching invoice data for client. try later")
	}

	return &models.NewlyGeneratedInvoiceData{
		StartDate: start,
		EndDate:   end,
		LineItems: slice.Map(summary.Projects, func(index int, project *models.SummaryItem) models.InvoiceLineItem {
			return models.InvoiceLineItem{
				Title:         project.Key,
				TotalSeconds:  int64(project.TotalFixed().Seconds()),
				AutoGenerated: true,
			}
		}),
	}, nil
}
