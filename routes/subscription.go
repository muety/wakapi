package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/services"
	"github.com/stripe/stripe-go/v74"
	portalSession "github.com/stripe/stripe-go/v74/billingportal/session"
	checkoutSession "github.com/stripe/stripe-go/v74/checkout/session"
	"net/http"
	"time"
)

type SubscriptionHandler struct {
	config     *conf.Config
	userSrvc   services.IUserService
	mailSrvc   services.IMailService
	httpClient *http.Client
}

func NewSubscriptionHandler(
	userService services.IUserService,
	mailService services.IMailService,
) *SubscriptionHandler {
	config := conf.Get()

	if config.Subscriptions.Enabled {
		stripe.Key = config.Subscriptions.StripeApiKey
	}

	return &SubscriptionHandler{
		config:     config,
		userSrvc:   userService,
		mailSrvc:   mailService,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// https://stripe.com/docs/billing/quickstart?lang=go

func (h *SubscriptionHandler) RegisterRoutes(router *mux.Router) {
	if !h.config.Subscriptions.Enabled {
		return
	}

	r := router.PathPrefix("/subscription").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).WithRedirectTarget(defaultErrorRedirectTarget()).Handler,
	)
	r.Path("/success").Methods(http.MethodGet).HandlerFunc(h.GetCheckoutSuccess)
	r.Path("/cancel").Methods(http.MethodGet).HandlerFunc(h.GetCheckoutCancel)
	r.Path("/checkout").Methods(http.MethodPost).HandlerFunc(h.PostCheckout)
	r.Path("/portal").Methods(http.MethodPost).HandlerFunc(h.PostPortal)

	router.Path("/webhook").Methods(http.MethodPost).HandlerFunc(h.PostWebhook)
}

func (h *SubscriptionHandler) PostCheckout(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	if user.Email == "" {
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription?error=%s", h.config.Server.BasePath, "missing e-mail address"), http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription?error=%s", h.config.Server.BasePath, "missing form values"), http.StatusFound)
		return
	}

	checkoutParams := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    &h.config.Subscriptions.StandardPriceId,
				Quantity: stripe.Int64(1),
			},
		},
		CustomerEmail: &user.Email,
		SuccessURL:    stripe.String(fmt.Sprintf("%s%s/subscription/success", h.config.Server.PublicUrl, h.config.Server.BasePath)),
		CancelURL:     stripe.String(fmt.Sprintf("%s%s/subscription/cacnel", h.config.Server.PublicUrl, h.config.Server.BasePath)),
	}

	session, err := checkoutSession.New(checkoutParams)
	if err != nil {
		conf.Log().Request(r).Error("failed to create stripe checkout session: %v", err)
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription?error=%s", h.config.Server.BasePath, "something went wrong"), http.StatusFound)
		return
	}

	http.Redirect(w, r, session.URL, http.StatusSeeOther)
}

func (h *SubscriptionHandler) PostPortal(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	if user.StripeCustomerId == "" {
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription?error=%s", h.config.Server.BasePath, "missing stripe customer reference, please contact us!"), http.StatusFound)
		return
	}

	portalParams := &stripe.BillingPortalSessionParams{
		Customer:  &user.StripeCustomerId,
		ReturnURL: &h.config.Server.PublicUrl,
	}

	session, err := portalSession.New(portalParams)
	if err != nil {
		conf.Log().Request(r).Error("failed to create stripe portal session: %v", err)
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription?error=%s", h.config.Server.BasePath, "something went wrong"), http.StatusFound)
		return
	}

	http.Redirect(w, r, session.URL, http.StatusSeeOther)
}

func (h *SubscriptionHandler) PostWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

func (h *SubscriptionHandler) GetCheckoutSuccess(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription?success=%s", h.config.Server.BasePath, "you have successfully subscribed to wakapi!"), http.StatusFound)
}

func (h *SubscriptionHandler) GetCheckoutCancel(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription", h.config.Server.BasePath), http.StatusFound)
}
