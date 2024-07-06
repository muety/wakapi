package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
)

type GoalService struct {
	config     *config.Config
	cache      *cache.Cache
	repository repositories.IGoalRepository
}

func NewGoalService(goalRepo repositories.IGoalRepository) *GoalService {
	return &GoalService{
		config:     config.Get(),
		cache:      cache.New(1*time.Hour, 2*time.Hour),
		repository: goalRepo,
	}
}

func (srv *GoalService) Create(newGoal *models.Goal) (*models.Goal, error) {
	return srv.repository.Create(newGoal)
}

func (srv *GoalService) Update(newGaol *models.Goal) (*models.Goal, error) {
	return srv.repository.Update(newGaol)
}

func (srv *GoalService) GetGoalForUser(id, userID string) (*models.Goal, error) {
	return srv.repository.GetByIdForUser(id, userID)
}

func (srv *GoalService) DeleteGoal(id, userID string) error {
	return srv.repository.DeleteByIdAndUser(id, userID)
}

func (srv *GoalService) FetchUserGoals(id string) ([]*models.Goal, error) {
	return srv.repository.FetchUserGoals(id)
}

func (srv *GoalService) LoadGoalChartData(goal *models.Goal, user *models.User, summarySrvc ISummaryService) ([]*models.GoalChartData, error) {
	rangeParam := "last_7_days"

	timezone := user.TZ()

	var start, end time.Time
	// range param takes precedence
	if err, parsedFrom, parsedTo := helpers.ResolveIntervalRawTZ(rangeParam, timezone); err == nil {
		start, end = parsedFrom, parsedTo
	} else {
		return nil, errors.New("invalid 'range' parameter")
	}

	// wakatime interprets end date as "inclusive", wakapi usually as "exclusive"
	// i.e. for wakatime, an interval 2021-04-29 - 2021-04-29 is actually 2021-04-29 - 2021-04-30,
	// while for wakapi it would be empty
	// see https://github.com/muety/wakapi/issues/192
	end = datetime.EndOfDay(end)

	if !end.After(start) {
		return nil, errors.New("'end' date must be after 'start' date")
	}

	overallParams := &models.SummaryParams{
		From: start,
		To:   end,
		User: user,
	}

	intervals := utils.SplitRangeByDays(overallParams.From, overallParams.To)

	filters := goal.GetGoalSummaryFilter()

	chartData := make([]*models.GoalChartData, len(intervals))
	for i, interval := range intervals {
		summary, err := summarySrvc.Aliased(interval[0], interval[1], user, summarySrvc.Retrieve, filters, end.After(time.Now()))
		if err != nil {
			return nil, err
		}
		// wakatime returns requested instead of actual summary range
		summary.FromTime = models.CustomTime(interval[0])
		summary.ToTime = models.CustomTime(interval[1].Add(-1 * time.Second))
		chartData[i] = prepareGoalSummary(summary, goal)
	}

	return chartData, nil
}

func formatDuration(seconds int64) string {
	switch {
	case seconds < 60:
		return fmt.Sprintf("%d secs", seconds)
	case seconds < 3600:
		minutes := seconds / 60
		return fmt.Sprintf("%d mins", minutes)
	default:
		hours := seconds / 3600
		remainingSeconds := seconds % 3600
		minutes := remainingSeconds / 60
		return fmt.Sprintf("%d hrs %d mins", hours, minutes)
	}
}

func moreOrLessText(diff int64) string {
	if diff > 0 {
		return "more"
	}
	return "less"
}

// TODO: Improve and actually implement
func getGoalStatus(goal *models.Goal, actualSeconds int64) (string, string) {
	actualDuration := formatDuration(actualSeconds)
	goalDuration := formatDuration(goal.Seconds)

	diff := actualSeconds - goal.Seconds
	diffDuration := formatDuration(diff)

	reason := fmt.Sprintf("You coded for %s, which is %s %s than your target of %s", actualDuration, diffDuration, moreOrLessText(diff), goalDuration)

	if diff >= 0 {
		return "success", reason
	}
	return "failed", reason
}

func prepareGoalSummary(s *models.Summary, goal *models.Goal) *models.GoalChartData {
	zone, _ := time.Now().Zone()
	total := s.TotalTime()

	status, statusText := getGoalStatus(goal, int64(total))

	// ComputeText
	goalRange := models.GoalChartRange{
		Date:     time.Now().Format(time.RFC3339),
		End:      s.ToTime.T(),
		Start:    s.FromTime.T(),
		Text:     "",
		Timezone: zone,
	}

	goalRange.ComputeText()

	return &models.GoalChartData{
		ActualSeconds:          total.Seconds(),
		GoalSeconds:            float64(goal.Seconds),
		RangeStatus:            status,     //write func to get range status
		RangeStatusReason:      statusText, //write func to get range status
		RangeStatusReasonShort: statusText,
		GoalSecondsText:        formatDuration(goal.Seconds),
		ActualSecondsText:      formatDuration(int64(total.Seconds())),
		Range:                  goalRange,
	}
}
