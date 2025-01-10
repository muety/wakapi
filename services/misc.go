package services

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"
	"go.uber.org/atomic"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/muety/wakapi/models"
)

const (
	countUsersEvery                  = 3 * time.Hour
	computeOldestDataEvery           = 6 * time.Hour
	notifyExpiringSubscriptionsEvery = 12 * time.Hour
)

const (
	notifyBeforeSubscriptionExpiry = 7 * 24 * time.Hour
)

var countLock = sync.Mutex{}
var firstDataLock = sync.Mutex{}

type MiscService struct {
	config           *config.Config
	userService      IUserService
	heartbeatService IHeartbeatService
	summaryService   ISummaryService
	keyValueService  IKeyValueService
	mailService      IMailService
	queueDefault     *artifex.Dispatcher
	queueWorkers     *artifex.Dispatcher
	queueMails       *artifex.Dispatcher
}

func NewMiscService(userService IUserService, heartbeatService IHeartbeatService, summaryService ISummaryService, keyValueService IKeyValueService, mailService IMailService) *MiscService {
	return &MiscService{
		config:           config.Get(),
		userService:      userService,
		heartbeatService: heartbeatService,
		summaryService:   summaryService,
		keyValueService:  keyValueService,
		mailService:      mailService,
		queueDefault:     config.GetDefaultQueue(),
		queueWorkers:     config.GetQueue(config.QueueProcessing),
		queueMails:       config.GetQueue(config.QueueMails),
	}
}

func (srv *MiscService) Schedule() {
	slog.Info("scheduling total time counting")
	if _, err := srv.queueDefault.DispatchEvery(srv.CountTotalTime, countUsersEvery); err != nil {
		config.Log().Error("failed to schedule user counting jobs", "error", err)
	}

	slog.Info("scheduling first data computing")
	if _, err := srv.queueDefault.DispatchEvery(srv.ComputeOldestHeartbeats, computeOldestDataEvery); err != nil {
		config.Log().Error("failed to schedule first data computing jobs", "error", err)
	}

	if srv.config.Subscriptions.Enabled && srv.config.Subscriptions.ExpiryNotifications && srv.config.App.DataRetentionMonths > 0 {
		slog.Info("scheduling subscription notifications")
		if _, err := srv.queueDefault.DispatchEvery(srv.NotifyExpiringSubscription, notifyExpiringSubscriptionsEvery); err != nil {
			config.Log().Error("failed to schedule subscription notification jobs", "error", err)
		}
	}

	// run once initially for a fresh instance
	if !srv.existsUsersTotalTime() {
		if err := srv.queueDefault.Dispatch(srv.CountTotalTime); err != nil {
			config.Log().Error("failed to dispatch user counting jobs", "error", err)
		}
	}
	if !srv.existsUsersFirstData() {
		if err := srv.queueDefault.Dispatch(srv.ComputeOldestHeartbeats); err != nil {
			config.Log().Error("failed to dispatch first data computing jobs", "error", err)
		}
	}
	if !srv.existsSubscriptionNotifications() && srv.config.Subscriptions.Enabled && srv.config.Subscriptions.ExpiryNotifications && srv.config.App.DataRetentionMonths > 0 {
		if err := srv.queueDefault.Dispatch(srv.NotifyExpiringSubscription); err != nil {
			config.Log().Error("failed to schedule subscription notification jobs", "error", err)
		}
	}
}

func (srv *MiscService) CountTotalTime() {
	slog.Info("counting users total time")
	if ok := countLock.TryLock(); !ok {
		config.Log().Warn("couldn't acquire lock for counting users total time, job is still pending")
	}
	defer countLock.Unlock()

	users, err := srv.userService.GetAll()
	if err != nil {
		config.Log().Error("failed to fetch users for time counting", "error", err)
		return
	}

	var totalTime = atomic.NewDuration(0)
	var pendingJobs sync.WaitGroup
	pendingJobs.Add(len(users))

	for _, u := range users {
		user := *u
		if err := srv.queueWorkers.Dispatch(func() {
			defer pendingJobs.Done()
			totalTime.Add(srv.countUserTotalTime(user.ID))
		}); err != nil {
			config.Log().Error("failed to enqueue counting job for user", "userID", user.ID)
			pendingJobs.Done()
		}
	}

	// persist
	go func(wg *sync.WaitGroup) {
		if !utils.WaitTimeout(&pendingJobs, 2*countUsersEvery) {
			if err := srv.keyValueService.PutString(&models.KeyStringValue{
				Key:   config.KeyLatestTotalTime,
				Value: totalTime.Load().String(),
			}); err != nil {
				config.Log().Error("failed to save total time count", "error", err)
			}

			if err := srv.keyValueService.PutString(&models.KeyStringValue{
				Key:   config.KeyLatestTotalUsers,
				Value: strconv.Itoa(len(users)),
			}); err != nil {
				config.Log().Error("failed to save total users count", "error", err)
			}
		} else {
			config.Log().Error("waiting for user counting jobs timed out")
		}
	}(&pendingJobs)
}

func (srv *MiscService) ComputeOldestHeartbeats() {
	slog.Info("computing users' first data")

	if err := srv.queueWorkers.Dispatch(func() {
		if ok := firstDataLock.TryLock(); !ok {
			config.Log().Warn("couldn't acquire lock for computing users' first data, job is still pending")
			return
		}
		defer firstDataLock.Unlock()

		results, err := srv.heartbeatService.GetFirstByUsers()
		if err != nil {
			config.Log().Error("failed to compute users' first data", "error", err)
			return
		}

		for _, entry := range results {
			if entry.Time.T().IsZero() {
				continue
			}

			kvKey := fmt.Sprintf("%s_%s", config.KeyFirstHeartbeat, entry.User)
			if err := srv.keyValueService.PutString(&models.KeyStringValue{
				Key:   kvKey,
				Value: entry.Time.T().Format(time.RFC822Z),
			}); err != nil {
				config.Log().Error("failed to save user's first heartbeat time", "error", err)
			}
		}
	}); err != nil {
		config.Log().Error("failed to enqueue computing first data for user", "error", err)
	}
}

// NotifyExpiringSubscription sends a reminder e-mail to all users, notifying them if their subscription has expired or is about to, given these conditions:
// - Data cleanup is enabled on the server (non-zero retention time)
// - Subscriptions are enabled on the server (aka. users can do something about their old data getting cleaned up)
// - User has an e-mail address configured
// - User's subscription has expired or is about to expire soon
// - User doesn't have upcoming auto-renewal (i.e. chose to cancel at some date in the future)
// - The user has gotten no such e-mail before recently
// Note: only one mail will be sent for either "expired" or "about to expire" state.
func (srv *MiscService) NotifyExpiringSubscription() {
	if srv.config.App.DataRetentionMonths <= 0 || !srv.config.Subscriptions.Enabled {
		return
	}

	now := time.Now()
	slog.Info("notifying users about soon to expire subscriptions")

	users, err := srv.userService.GetAll()
	if err != nil {
		config.Log().Error("failed to fetch users for subscription notifications", "error", err)
		return
	}

	var subscriptionReminders map[string][]*models.KeyStringValue
	if result, err := srv.keyValueService.GetByPrefix(config.KeySubscriptionNotificationSent); err == nil {
		subscriptionReminders = slice.GroupWith[*models.KeyStringValue, string](result, func(kv *models.KeyStringValue) string {
			return strings.Replace(kv.Key, config.KeySubscriptionNotificationSent+"_", "", 1)
		})
	} else {
		config.Log().Error("failed to fetch key-values for subscription notifications", "error", err)
		return
	}

	for _, u := range users {
		if u.HasActiveSubscription() && u.Email == "" {
			config.Log().Warn("invalid state: user has active subscription but no e-mail address set", "userID", u.ID)
		}

		var alreadySent bool
		if kvs, ok := subscriptionReminders[u.ID]; ok && len(kvs) > 0 {
			// don't send again while subscription hasn't had chance to have been renewed
			if sendDate, err := time.Parse(time.RFC822Z, kvs[0].Value); err == nil && now.Sub(sendDate) <= notifyBeforeSubscriptionExpiry {
				alreadySent = true
			} else if err != nil {
				config.Log().Error("failed to parse date for last sent subscription notification mail", "userID", u.ID, "error", err)
				alreadySent = true
			}
		}

		// skip users without e-mail address
		// skip users who already received a notification before
		// skip users who either never had a subscription before or intentionally deleted it
		// skip users who have upcoming auto-renewal (everyone except users who chose to cancel subscription at later date)
		if alreadySent || u.Email == "" || u.SubscribedUntil == nil || (u.SubscriptionRenewal != nil && u.SubscriptionRenewal.T().After(now)) {
			continue
		}

		expired, expiredSince := u.SubscriptionExpiredSince()
		if expired || (expiredSince < 0 && expiredSince*-1 <= notifyBeforeSubscriptionExpiry) {
			srv.sendSubscriptionNotificationScheduled(u, expired)
		}
	}
}

func (srv *MiscService) countUserTotalTime(userId string) time.Duration {
	result, err := srv.summaryService.Aliased(time.Time{}, time.Now(), &models.User{ID: userId}, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		config.Log().Error("failed to count total for user", "userID", userId, "error", err)
		return 0
	}
	return result.TotalTime()
}

func (srv *MiscService) sendSubscriptionNotificationScheduled(user *models.User, hasExpired bool) {
	u := *user
	srv.queueMails.Dispatch(func() {
		slog.Info("sending subscription expiry notification mail", "userID", u.ID, "expired", hasExpired)
		defer time.Sleep(10 * time.Second)

		if err := srv.mailService.SendSubscriptionNotification(&u, hasExpired); err != nil {
			config.Log().Error("failed to send subscription notification mail to user", "userID", u.ID, "error", err)
			return
		}

		if err := srv.keyValueService.PutString(&models.KeyStringValue{
			Key:   fmt.Sprintf("%s_%s", config.KeySubscriptionNotificationSent, u.ID),
			Value: time.Now().Format(time.RFC822Z),
		}); err != nil {
			config.Log().Error("failed to update subscription notification status key-value for user", "userID", u.ID, "error", err)
		}
	})
}

func (srv *MiscService) existsUsersTotalTime() bool {
	results, err := srv.keyValueService.GetByPrefix(config.KeyLatestTotalTime)
	if err != nil {
		config.Log().Error("failed to fetch latest time key-values", "error", err)
	}
	return len(results) > 0
}

func (srv *MiscService) existsUsersFirstData() bool {
	results, err := srv.keyValueService.GetByPrefix(config.KeyFirstHeartbeat)
	if err != nil {
		config.Log().Error("failed to fetch first heartbeats key-values", "error", err)
	}
	return len(results) > 0
}

func (srv *MiscService) existsSubscriptionNotifications() bool {
	results, err := srv.keyValueService.GetByPrefix(config.KeySubscriptionNotificationSent)
	if err != nil {
		config.Log().Error("failed to fetch notifications key-values", "error", err)
	}
	return len(results) > 0
}
