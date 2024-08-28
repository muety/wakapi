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

type ClientsApiHandler struct {
	db            *gorm.DB
	config        *conf.Config
	clientService services.IClientService
	userService   services.IUserService
}

func NewClientsApiHandler(db *gorm.DB, clientService services.IClientService, userService services.IUserService) *ClientsApiHandler {
	return &ClientsApiHandler{
		db:            db,
		clientService: clientService,
		userService:   userService,
		config:        conf.Get(),
	}
}

func (h *ClientsApiHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userService).Handler)
		r.Post("/compat/wakatime/v1/users/{user}/clients", h.Create)
		r.Get("/compat/wakatime/v1/users/{user}/clients", h.FetchUserClients)
		r.Get("/compat/wakatime/v1/users/{user}/clients/{id}", h.GetClient)
		r.Put("/compat/wakatime/v1/users/{user}/clients/{id}", h.UpdateClient)
		r.Delete("/compat/wakatime/v1/users/{user}/clients/{id}", h.DeleteClient)
	})
}

func (h *ClientsApiHandler) Ping(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": "pong",
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (h *ClientsApiHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	clientID := chi.URLParam(r, "id")

	if clientID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
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

	client, err := h.clientService.GetClientForUser(clientID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Client Cannot Be Found",
			"status":  http.StatusNotFound,
		})
		return
	}

	_, err = h.clientService.Update(client, params)
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

func (h *ClientsApiHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	clientID := chi.URLParam(r, "id")

	if clientID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	client, err := h.clientService.GetClientForUser(clientID, user.ID)
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

func (h *ClientsApiHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	clientID := chi.URLParam(r, "id")

	if clientID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	err := h.clientService.DeleteClient(clientID, user.ID)
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

func (h *ClientsApiHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	_, err = h.clientService.Create(params)
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

func (h *ClientsApiHandler) FetchUserClients(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	query := r.URL.Query().Get("q")

	clients, err := h.clientService.FetchUserClients(user.ID, query)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Error fetching clients. Try later.",
		})
		return
	}
	response := map[string]interface{}{
		"data": clients,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}
