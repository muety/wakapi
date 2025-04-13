package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	uuid "github.com/satori/go.uuid"
)

func (a *APIv1) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
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

	invoice, err := a.services.Invoice().GetInvoiceForUser(InvoiceID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invoice Cannot Be Found",
			"status":  http.StatusNotFound,
		})
		return
	}

	_, err = a.services.Invoice().Update(invoice, params)
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

func (a *APIv1) GetInvoice(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	InvoiceID := chi.URLParam(r, "id")

	if InvoiceID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	Invoice, err := a.services.Invoice().GetInvoiceForUser(InvoiceID, user.ID)
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

func (a *APIv1) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	invoiceID := chi.URLParam(r, "id")

	if invoiceID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	err := a.services.Invoice().DeleteInvoice(invoiceID, user.ID)
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

func (a *APIv1) CreateInvoice(w http.ResponseWriter, r *http.Request) {
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

	client, err := a.services.Client().GetClientForUser(params.ClientID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Client Cannot Be Found",
			"status":  http.StatusNotFound,
		})
		return
	}

	generatedData, err := a.FetchInvoiceLineItems(user, client, params.StartDate, params.EndDate)
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
		Origin:         fmt.Sprintf("%s \nFreelancer", user.Email),
		Tax:            0,
	}

	_, err = a.services.Invoice().Create(newInvoice)
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

func (a *APIv1) FetchUserInvoices(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	query := r.URL.Query().Get("q")

	Invoices, err := a.services.Invoice().FetchUserInvoices(user.ID, query)
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

func (a *APIv1) FetchInvoiceLineItems(user *models.User, client *models.Client, start_date string, end_date string) (*models.NewlyGeneratedInvoiceData, error) {
	start, err := helpers.ParseDateTimeTZ(start_date, user.TZ())
	if err != nil {
		return nil, fmt.Errorf("invalid date %s provided", start_date)
	}

	end, err := helpers.ParseDateTimeTZ(end_date, user.TZ())
	if err != nil {
		return nil, fmt.Errorf("invalid date %s provided", end_date)
	}

	summary, err := a.services.Client().FetchClientInvoiceLineItems(client, user, a.services.Summary(), start, end)
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
