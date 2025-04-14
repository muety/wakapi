package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

type GoalService struct {
	config *config.Config
	db     *gorm.DB
}

func NewGoalService(db *gorm.DB) *GoalService {
	return &GoalService{
		config: config.Get(),
		db:     db,
	}
}

func (srv *GoalService) Create(newGoal *models.Goal) (*models.Goal, error) {
	result := srv.db.Create(newGoal)
	if err := result.Error; err != nil {
		return nil, err
	}
	return newGoal, nil
}

func (srv *GoalService) Update(newGaol *models.Goal) (*models.Goal, error) {
	updateMap := map[string]interface{}{
		"custom_title": newGaol.CustomTitle,
	}

	result := srv.db.Model(newGaol).Updates(updateMap)
	if err := result.Error; err != nil {
		return nil, err
	}

	return newGaol, nil
}

func (srv *GoalService) GetGoalForUser(goalID, userID string) (*models.Goal, error) {
	g := &models.Goal{}

	err := srv.db.Where(models.Goal{ID: goalID, UserID: userID}).First(g).Error
	if err != nil {
		return g, err
	}

	if g.ID != "" {
		return g, nil
	}
	return nil, err
}

func (srv *GoalService) DeleteGoal(goalId, userID string) error {
	if err := srv.db.
		Where("id = ?", goalId).
		Where("user_id = ?", userID).
		Delete(models.Goal{}).Error; err != nil {
		return err
	}
	return nil
}

func (srv *GoalService) FetchUserGoals(userID string) ([]*models.Goal, error) {
	var goals []*models.Goal
	if err := srv.db.
		Order("created_at desc").
		Limit(100). // TODO: paginate
		Where(&models.Goal{UserID: userID}).
		Find(&goals).Error; err != nil {
		return nil, err
	}
	return goals, nil
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

type IGoalService interface {
	Create(*models.Goal) (*models.Goal, error)
	GetGoalForUser(id, userID string) (*models.Goal, error)
	Update(newGoal *models.Goal) (*models.Goal, error)
	DeleteGoal(id string, userID string) error
	FetchUserGoals(id string) ([]*models.Goal, error)
	LoadGoalChartData(goal *models.Goal, user *models.User, summarySrvc ISummaryService) ([]*models.GoalChartData, error)
}
