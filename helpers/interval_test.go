package helpers

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestResolveMaximumRange_Default(t *testing.T) {
	for i := 1; i <= 366; i++ {
		err1, maximumInterval := ResolveMaximumRange(i)
		err2, from, to := ResolveIntervalTZ(maximumInterval, time.UTC)

		assert.Nil(t, err1)
		assert.Nil(t, err2)
		assert.LessOrEqual(t, to.Sub(from), time.Duration(i*24)*time.Hour)
	}
}

func TestResolveMaximumRange_EdgeCases(t *testing.T) {
	err, _ := ResolveMaximumRange(0)
	assert.NotNil(t, err)

	_, maximumInterval := ResolveMaximumRange(-1)
	assert.Equal(t, models.IntervalAny, maximumInterval)
}
