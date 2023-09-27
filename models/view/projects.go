package view

import (
	"fmt"
	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"image/color"
)

type ProjectsViewModel struct {
	Messages
	User       *models.User
	Projects   []*models.ProjectStats
	ApiKey     string
	PageParams *utils.PageParams
	maxCount   int64
}

func (s *ProjectsViewModel) LangIcon(lang string) string {
	return GetLanguageIcon(lang)
}

func (s *ProjectsViewModel) BackgroundIntensity(idx int) string {
	maxCount := s.getMaxCount()
	intensity := float64(s.Projects[idx].Count) / float64(maxCount)
	return fadeColorToTransparent("#047857", intensity)
}

func (s *ProjectsViewModel) WithSuccess(m string) *ProjectsViewModel {
	s.SetSuccess(m)
	return s
}

func (s *ProjectsViewModel) WithError(m string) *ProjectsViewModel {
	s.SetError(m)
	return s
}

func (s *ProjectsViewModel) getMaxCount() int64 {
	if s.maxCount == 0 {
		s.maxCount = mathutil.Max(slice.Map[*models.ProjectStats, int64](s.Projects, func(i int, p *models.ProjectStats) int64 {
			return p.Count
		})...)
	}
	return mathutil.Max(s.maxCount, 1)
}

func fadeColorToTransparent(colorHex string, transparency float64) string {
	left := utils.ParseHexColor(colorHex)
	right := &color.RGBA{R: left.R, G: left.G, B: left.B, A: uint8(transparency * 255)}
	return fmt.Sprintf("background: transparent; background: linear-gradient(90deg, rgba(%d, %d, %d, 0) 0%%, rgba(%d, %d, %d, 0) 50%%, rgba(%d, %d, %d, %.2f) 100%%);",
		left.R, left.G, left.B,
		left.R, left.G, left.B,
		right.R, right.G, right.B, float32(right.A)/255,
	)
}
