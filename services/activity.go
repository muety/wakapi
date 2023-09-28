package services

import (
	"bytes"
	"errors"
	"fmt"
	svg "github.com/ajstarks/svgo/float"
	"github.com/alitto/pond"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
	"math"
	"sync"
	"time"
)

const (
	gridRows   = 7
	cellWidth  = 20
	cellHeight = 20
	colorMin   = "#dce3e1"
	colorMax   = "#047857"
)

type ActivityService struct {
	config         *config.Config
	cache          *cache.Cache
	summaryService ISummaryService
}

func NewActivityService(summaryService ISummaryService) *ActivityService {
	return &ActivityService{
		config:         config.Get(),
		cache:          cache.New(6*time.Hour, 6*time.Hour),
		summaryService: summaryService,
	}
}

// GetChart generates an activity chart for a given user and the given time interval, similar to GitHub's contribution timeline. See https://github.com/muety/wakapi/issues/12.
// Please note: currently, only yearly charts ("last_12_months") are supported. However, we could fairly easily restructure this to support dynamic intervals.
func (s *ActivityService) GetChart(user *models.User, interval *models.IntervalKey, skipCache bool) (string, error) {
	cacheKey := fmt.Sprintf("chart_%s_%s", user.ID, (*interval)[0])
	if result, found := s.cache.Get(cacheKey); found && !skipCache {
		return result.(string), nil
	}

	switch interval {
	case models.IntervalPast12Months:
		chart, err := s.getChartPastYear(user)
		if err == nil {
			s.cache.SetDefault(cacheKey, chart) // TODO: cache compressed?
		}
		return chart, err
	default:
		return "", errors.New("unsupported interval")
	}
}

func (s *ActivityService) getChartPastYear(user *models.User) (string, error) {
	err, from, to := helpers.ResolveIntervalTZ(models.IntervalPast12Months, user.TZ())
	from = datetime.BeginOfWeek(from, time.Monday)
	if err != nil {
		return "", err
	}

	intervals := utils.SplitRangeByDays(from, to)
	summaries := make([]*models.Summary, len(intervals))

	wp := pond.New(utils.HalfCPUs(), 0)
	mut := sync.RWMutex{}

	// fetch summaries
	for i, interval := range intervals {
		i := i // https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable
		interval := interval

		wp.Submit(func() {
			summary, err := s.summaryService.Retrieve(interval[0], interval[1], user, nil)
			fmt.Println(summary == nil)
			if err != nil {
				config.Log().Warn("failed to retrieve summary for '%s' between %v and %v for activity chart", user.ID, from, to)
				summary = models.NewEmptySummary()
				summary.FromTime = models.CustomTime(from)
				summary.ToTime = models.CustomTime(to)
				summary.UserID = user.ID
				summary.User = user
			}
			mut.Lock()
			summaries[i] = summary
			mut.Unlock()
		})
	}

	wp.StopAndWait()

	maxTotal := models.Summaries(summaries).MaxTotalTime()

	var (
		colorRGBAMin = utils.HexToRGBA(colorMin)
		colorRGBAMax = utils.HexToRGBA(colorMax)
	)

	// generate svg
	buf := &bytes.Buffer{}
	canvas := svg.New(buf)
	canvas.Start(math.Ceil(float64(len(summaries))/float64(gridRows))*cellWidth, gridRows*cellHeight)
	for i, s := range summaries {
		total := s.TotalTime()
		fillColor := utils.RGBAToHex(utils.FadeColors(colorRGBAMin, colorRGBAMax, float64(total)/float64(maxTotal)))
		canvas.Group()
		canvas.Title(fmt.Sprintf("%s on %s", helpers.FmtWakatimeDuration(total), helpers.FormatDateHuman(s.FromTime.T())))
		canvas.Rect(float64(i/gridRows)*cellWidth, float64((i%gridRows)*cellHeight), cellWidth, cellHeight, fmt.Sprintf("fill: %s; fill-opacity: 1; stroke: #fff; stroke-width: 1; stroke-linecap: square; stroke-opacity: 1", fillColor))
		canvas.Gend()
	}
	canvas.End()

	return buf.String(), nil
}
