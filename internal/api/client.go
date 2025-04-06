package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	uuid "github.com/satori/go.uuid"
)

func (a *APIv1) UpdateClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	clientID := chi.URLParam(r, "id")

	if clientID == "" {
		response := map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		}
		sendJSON(
			w,
			http.StatusBadRequest,
			response,
			"No client Id provided",
			"",
		)
		return
	}

	var params = &models.ClientUpdate{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid Input",
			"status":  http.StatusBadRequest,
		})
	}

	client, err := a.services.Client().GetClientForUser(clientID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Client Cannot Be Found",
			"status":  http.StatusNotFound,
		})
		return
	}

	_, err = a.services.Client().Update(client, params)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "Error updating client",
			"error_message": err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": client,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) GetClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	clientID := chi.URLParam(r, "id")

	if clientID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	client, err := a.services.Client().GetClientForUser(clientID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Client Cannot Be Found",
			"status":  http.StatusBadRequest,
		})
		return
	}
	response := map[string]interface{}{
		"data": client,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) DeleteClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	clientID := chi.URLParam(r, "id")

	if clientID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	err := a.services.Client().DeleteClient(clientID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Client Cannot Be Deleted",
			"status":  http.StatusBadRequest,
		})
		return
	}
	response := map[string]interface{}{
		"message": "Client deleted successfully",
	}
	helpers.RespondJSON(w, r, http.StatusAccepted, response)
}

func (a *APIv1) CreateClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)

	var params = &models.Client{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil || params.Name == "" || params.Projects == nil || len(params.Projects) == 0 {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid Input: Name, Projects are required",
			"status":  http.StatusBadRequest,
		})
		return
	}

	params.UserID = user.ID
	params.ID = uuid.NewV4().String()

	_, err = a.services.Client().Create(params)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "An unexpected error occurred. Try again later",
			"error_message": err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": params,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) FetchUserClients(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	query := r.URL.Query().Get("q")

	clients, err := a.services.Client().FetchUserClients(user.ID, query)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Error fetching clients. Try later.",
			"error":   err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": clients,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) FetchInvoiceLineItemsForClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	clientID := chi.URLParam(r, "id")

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "from and to fields are required for the invoice duration",
			"status":  http.StatusBadRequest,
		})
		return
	}

	client, err := a.services.Client().GetClientForUser(clientID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusNotFound, map[string]interface{}{
			"message": "Client cannot be found",
			"status":  http.StatusBadRequest,
		})
		return
	}

	start, err := helpers.ParseDateTimeTZ(from, user.TZ())
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": fmt.Sprintf("Invalid date %s provided", from),
			"status":  http.StatusBadRequest,
		})
		return
	}

	end, err := helpers.ParseDateTimeTZ(to, user.TZ())
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": fmt.Sprintf("Invalid date %s provided", end),
			"status":  http.StatusBadRequest,
		})
		return
	}

	summary, err := a.services.Client().FetchClientInvoiceLineItems(client, user, a.services.Summary(), start, end)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
			"message": "Error fetching invoice data for client. Try later.",
			"error":   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"client":     client,
		"line_items": summary.Projects,
	}
	sendJSONSuccess(w, http.StatusAccepted, response)
}
