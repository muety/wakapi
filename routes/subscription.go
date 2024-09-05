package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/leandro-lugaresi/hub"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/stripe/stripe-go/v74"
	stripePortalSession "github.com/stripe/stripe-go/v74/billingportal/session"
	stripeCheckoutSession "github.com/stripe/stripe-go/v74/checkout/session"
	stripeCustomer "github.com/stripe/stripe-go/v74/customer"
	stripePrice "github.com/stripe/stripe-go/v74/price"
	stripeSubscription "github.com/stripe/stripe-go/v74/subscription"
	"github.com/stripe/stripe-go/v74/webhook"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

/*
  How to integrate with Stripe?
  ---
  1. Create a plan with recurring payment (https://dashboard.stripe.com/test/products?active=true), copy its ID and save it as 'standard_price_id'
  2. Create a webhook (https://dashboard.stripe.com/test/webhooks), with target URL '/subscription/webhook' and events ['customer.subscription.created', 'customer.subscription.updated', 'customer.subscription.deleted', 'checkout.session.completed'], copy the endpoint secret and save it to 'stripe_endpoint_secret'
  3. Create a secret API key (https://dashboard.stripe.com/test/apikeys), copy it and save it to 'stripe_secret_key'
  4. Copy the publishable API key (https://dashboard.stripe.com/test/apikeys) and save it to 'stripe_api_key'
*/

// TODO: move all logic inside this controller into a separate service

type SubscriptionHandler struct {
	config       *conf.Config
	eventBus     *hub.Hub
	userSrvc     services.IUserService
	mailSrvc     services.IMailService
	keyValueSrvc services.IKeyValueService
	httpClient   *http.Client
}

func NewSubscriptionHandler(
	userService services.IUserService,
	mailService services.IMailService,
	keyValueService services.IKeyValueService,
) *SubscriptionHandler {
	config := conf.Get()
	eventBus := conf.EventBus()

	if config.Subscriptions.Enabled {
		stripe.Key = config.Subscriptions.StripeSecretKey

		price, err := stripePrice.Get(config.Subscriptions.StandardPriceId, nil)
		if err != nil {
			conf.Log().Fatal("failed to fetch stripe plan details", "error", err)
		}
		config.Subscriptions.StandardPrice = strings.TrimSpace(fmt.Sprintf("%2.f €", price.UnitAmountDecimal/100.0)) // TODO: respect actual currency

		slog.Info("enabling subscriptions with stripe payment", "price", config.Subscriptions.StandardPrice)
	}

	handler := &SubscriptionHandler{
		config:       config,
		userSrvc:     userService,
		mailSrvc:     mailService,
		keyValueSrvc: keyValueService,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}

	onUserDelete := eventBus.Subscribe(0, conf.EventUserDelete)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			user := m.Fields[conf.FieldPayload].(*models.User)
			if !user.HasActiveSubscription() {
				continue
			}

			slog.Info("cancelling subscription for user upon account deletion", "userID", user.ID, "email", user.Email, "stripeCustomerID", user.StripeCustomerId)
			if err := handler.cancelUserSubscription(user); err == nil {
				slog.Info("successfully cancelled subscription for user", "userID", user.ID, "email", user.Email, "stripeCustomerID", user.StripeCustomerId)
			} else {
				conf.Log().Error("failed to cancel subscription for user", "userID", user.ID, "email", user.Email, "stripeCustomerID", user.StripeCustomerId, "error", err)
			}
		}
	}(&onUserDelete)

	return handler
}

// https://stripe.com/docs/billing/quickstart?lang=go

func (h *SubscriptionHandler) RegisterRoutes(router chi.Router) {
	if !h.config.Subscriptions.Enabled {
		return
	}

	subRouterPublic := chi.NewRouter()
	subRouterPublic.Get("/success", h.GetCheckoutSuccess)
	subRouterPublic.Get("/cancel", h.GetCheckoutCancel)
	subRouterPublic.Post("/webhook", h.PostWebhook)

	subRouterPrivate := chi.NewRouter()
	subRouterPrivate.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).
			WithRedirectTarget(defaultErrorRedirectTarget()).
			WithRedirectErrorMessage("unauthorized").Handler,
	)
	subRouterPrivate.Post("/checkout", h.PostCheckout)
	subRouterPrivate.Post("/portal", h.PostPortal)

	subRouterPublic.Mount("/", subRouterPrivate)
	router.Mount("/subscription", subRouterPublic)
}

func (h *SubscriptionHandler) PostCheckout(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	user := middlewares.GetPrincipal(r)
	if user.Email == "" {
		routeutils.SetError(r, w, "missing e-mail address")
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription", h.config.Server.BasePath), http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		routeutils.SetError(r, w, "missing form values")
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription", h.config.Server.BasePath), http.StatusFound)
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
		ClientReferenceID:   &user.ID,
		AllowPromotionCodes: stripe.Bool(true),
		SuccessURL:          stripe.String(fmt.Sprintf("%s%s/subscription/success", h.config.Server.PublicUrl, h.config.Server.BasePath)),
		CancelURL:           stripe.String(fmt.Sprintf("%s%s/subscription/cancel", h.config.Server.PublicUrl, h.config.Server.BasePath)),
	}

	if user.StripeCustomerId != "" {
		checkoutParams.Customer = &user.StripeCustomerId
	} else {
		checkoutParams.CustomerEmail = &user.Email
	}

	session, err := stripeCheckoutSession.New(checkoutParams)
	if err != nil {
		conf.Log().Request(r).Error("failed to create stripe checkout session", "error", err)
		routeutils.SetError(r, w, "something went wrong")
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription", h.config.Server.BasePath), http.StatusFound)
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
		routeutils.SetError(r, w, "no subscription found with your e-mail address, please contact us!")
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription", h.config.Server.BasePath), http.StatusFound)
		return
	}

	portalParams := &stripe.BillingPortalSessionParams{
		Customer:  &user.StripeCustomerId,
		ReturnURL: &h.config.Server.PublicUrl,
	}

	session, err := stripePortalSession.New(portalParams)
	if err != nil {
		conf.Log().Request(r).Error("failed to create stripe portal session", "error", err)
		routeutils.SetError(r, w, "something went wrong")
		http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription", h.config.Server.BasePath), http.StatusFound)
		return
	}

	http.Redirect(w, r, session.URL, http.StatusSeeOther)
}

func (h *SubscriptionHandler) PostWebhook(w http.ResponseWriter, r *http.Request) {
	bodyReader := http.MaxBytesReader(w, r.Body, int64(65536))
	payload, err := io.ReadAll(bodyReader)
	if err != nil {
		conf.Log().Request(r).Error("error in stripe webhook request", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEventWithOptions(payload, r.Header.Get("Stripe-Signature"), h.config.Subscriptions.StripeEndpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		conf.Log().Request(r).Error("stripe webhook signature verification failed", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "customer.subscription.deleted",
		"customer.subscription.updated",
		"customer.subscription.created":
		// example payload: https://pastr.de/p/k7bx3alx38b1iawo6amtx09k
		subscription, err := h.parseSubscriptionEvent(w, r, event)
		if err != nil {
			return // status code already written
		}
		slog.Info("received stripe subscription event", "eventType", event.Type, "subscriptionID", subscription.ID, "customerID", subscription.Customer.ID)

		// first, try to get user by associated customer id (requires checkout.session.completed event to have been processed before)
		user, err := h.userSrvc.GetUserByStripeCustomerId(subscription.Customer.ID)
		if err != nil {
			conf.Log().Request(r).Warn("failed to find user with stripe customer id to update their subscription", "customerID", subscription.Customer.ID, "status", subscription.Status)

			// second, resolve customer and try to get user by email
			customer, err := stripeCustomer.Get(subscription.Customer.ID, nil)
			if err != nil {
				conf.Log().Request(r).Error("failed to fetch stripe customer", "customerID", subscription.Customer.ID, "error", err)
				w.WriteHeader(http.StatusOK) // don't make stripe retry the event
				return
			}

			u, err := h.userSrvc.GetUserByEmail(customer.Email)
			if err != nil {
				conf.Log().Request(r).Error("failed to get user for processing subscription event", "email", customer.Email, "customerID", subscription.Customer.ID, "subscriptionID", subscription.ID, "error", err)
				w.WriteHeader(http.StatusOK) // don't make stripe retry the event
				return
			}
			user = u
		}

		if err := h.handleSubscriptionEvent(subscription, user); err != nil {
			conf.Log().Request(r).Error("failed to handle subscription event", "eventID", event.ID, "eventType", event.Type, "userID", user.ID, "error", err)
			w.WriteHeader(http.StatusOK) // don't make stripe retry the event
			return
		}

	case "checkout.session.completed":
		// example payload: https://pastr.de/p/d01iniw9naq9hkmvyqtxin2w
		checkoutSession, err := h.parseCheckoutSessionEvent(w, r, event)
		if err != nil {
			return // status code already written
		}
		slog.Info("received stripe checkout session event", "eventType", event.Type, "sessionID", checkoutSession.ID, "customerID", checkoutSession.Customer.ID, "customerEmail", checkoutSession.CustomerEmail)

		user, err := h.userSrvc.GetUserById(checkoutSession.ClientReferenceID)
		if err != nil {
			conf.Log().Request(r).Error("failed to find user to update associated stripe customer", "userID", user.ID, "customerID", checkoutSession.Customer.ID)
			return // status code already written
		}

		if user.StripeCustomerId == "" {
			user.StripeCustomerId = checkoutSession.Customer.ID
			if _, err := h.userSrvc.Update(user); err != nil {
				conf.Log().Request(r).Error("failed to update stripe customer id for user", "customerID", checkoutSession.Customer.ID, "userID", user.ID, "error", err)
			} else {
				slog.Info("associated user with stripe customer", "userID", user.ID, "stripeCustomerID", checkoutSession.Customer.ID)
			}
		} else if user.StripeCustomerId != checkoutSession.Customer.ID {
			conf.Log().Request(r).Error("invalid state: tried to associate user with stripe customer, but customer already assigned", "userID", user.ID, "newCustomerID", checkoutSession.Customer.ID, "existingCustomerID", user.StripeCustomerId)
		}

	default:
		slog.Warn("got stripe event with no handler defined", "eventType", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SubscriptionHandler) GetCheckoutSuccess(w http.ResponseWriter, r *http.Request) {
	routeutils.SetSuccess(r, w, "you have successfully subscribed to Wakapi!")
	http.Redirect(w, r, fmt.Sprintf("%s/settings", h.config.Server.BasePath), http.StatusFound)
}

func (h *SubscriptionHandler) GetCheckoutCancel(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("%s/settings#subscription", h.config.Server.BasePath), http.StatusFound)
}

func (h *SubscriptionHandler) handleSubscriptionEvent(subscription *stripe.Subscription, user *models.User) error {
	var hasSubscribed bool

	switch subscription.Status {
	case "active":
		until := models.CustomTime(time.Unix(subscription.CurrentPeriodEnd, 0))

		if user.SubscribedUntil == nil || !user.SubscribedUntil.T().Equal(until.T()) {
			hasSubscribed = true
			user.SubscribedUntil = &until
			user.SubscriptionRenewal = &until
			slog.Info("user got active subscription", "userID", user.ID, "subscriptionID", subscription.ID, "subscribedUntil", user.SubscribedUntil)
		}

		if cancelAt := time.Unix(subscription.CancelAt, 0); !cancelAt.IsZero() && cancelAt.After(time.Now()) {
			user.SubscriptionRenewal = nil
			slog.Info("user chose to cancel subscription", "userID", user.ID, "subscriptionID", subscription.ID, "cancelAt", cancelAt)
		}
	case "canceled", "unpaid", "incomplete_expired":
		user.SubscribedUntil = nil
		user.SubscriptionRenewal = nil
		slog.Info("user's subscription got canceled due to status update", "userID", user.ID, "subscriptionID", subscription.ID, "status", subscription.Status)
	default:
		slog.Info("got subscription status update", "subscriptionID", subscription.ID, "status", subscription.Status, "userID", user.ID)
		return nil
	}

	_, err := h.userSrvc.Update(user)
	if err == nil && hasSubscribed {
		go h.clearSubscriptionNotificationStatus(user.ID)
	}
	return err
}

func (h *SubscriptionHandler) parseSubscriptionEvent(w http.ResponseWriter, r *http.Request, event stripe.Event) (*stripe.Subscription, error) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		conf.Log().Request(r).Error("failed to parse stripe webhook payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return &subscription, nil
}

func (h *SubscriptionHandler) parseCheckoutSessionEvent(w http.ResponseWriter, r *http.Request, event stripe.Event) (*stripe.CheckoutSession, error) {
	var checkoutSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
		conf.Log().Request(r).Error("failed to parse stripe webhook payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	return &checkoutSession, nil
}

func (h *SubscriptionHandler) cancelUserSubscription(user *models.User) error {
	// TODO: directly store subscription id with user object
	subscription, err := h.findCurrentStripeSubscription(user.StripeCustomerId)
	if err != nil {
		return err
	}
	_, err = stripeSubscription.Cancel(subscription.ID, nil)
	return err
}

func (h *SubscriptionHandler) findStripeCustomerByEmail(email string) (*stripe.Customer, error) {
	params := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf(`email:"%s"`, email),
		},
	}

	results := stripeCustomer.Search(params)
	if err := results.Err(); err != nil {
		return nil, err
	}

	if results.Next() {
		return results.Customer(), nil
	} else {
		return nil, errors.New("no customer found with given criteria")
	}
}

func (h *SubscriptionHandler) findCurrentStripeSubscription(customerId string) (*stripe.Subscription, error) {
	paramStatus := "active"
	params := &stripe.SubscriptionListParams{
		Customer: &customerId,
		Price:    &h.config.Subscriptions.StandardPriceId,
		Status:   &paramStatus,
		CurrentPeriodEndRange: &stripe.RangeQueryParams{
			GreaterThan: time.Now().Unix(),
		},
	}
	params.Filters.AddFilter("limit", "", "1")

	if result := stripeSubscription.List(params); result.Next() {
		return result.Subscription(), nil
	}
	return nil, fmt.Errorf("no active subscription found for customer '%s'", customerId)
}

func (h *SubscriptionHandler) clearSubscriptionNotificationStatus(userId string) {
	key := fmt.Sprintf("%s_%s", conf.KeySubscriptionNotificationSent, userId)
	if err := h.keyValueSrvc.DeleteString(key); err != nil {
		slog.Warn("failed to delete", "key", key, "error", err)
	}
}
