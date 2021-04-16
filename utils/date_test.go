package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDate_Ceil(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			"02 Jan 06 15:04 MST",
			"03 Jan 06 00:00 MST",
		},
		{
			"03 Jan 06 00:00 MST",
			"03 Jan 06 00:00 MST",
		},
	}

	for _, test := range tests {
		inDate, _ := time.Parse(time.RFC822, test.in)
		outDate, _ := time.Parse(time.RFC822, test.out)
		out := CeilDate(inDate)
		assert.Equal(t, outDate, out)
	}
}
