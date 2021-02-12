package metrics

import (
	"fmt"
	"strings"
)

// Hand-crafted Prometheus metrics
// Since we're only using very simple counters in this application,
// we don't actually need the official client SDK as a dependency

type Metrics []Metric

func (m Metrics) Print() (output string) {
	printedMetrics := make(map[string]bool)
	for _, m := range m {
		if _, ok := printedMetrics[m.Key()]; !ok {
			output += fmt.Sprintf("%s\n", m.Header())
			printedMetrics[m.Key()] = true
		}
		output += fmt.Sprintf("%s\n", m.Print())
	}

	return output
}

func (m Metrics) Len() int {
	return len(m)
}

func (m Metrics) Less(i, j int) bool {
	return strings.Compare(m[i].Key(), m[j].Key()) < 0
}

func (m Metrics) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type Metric interface {
	Key() string
	Header() string
	Print() string
}
