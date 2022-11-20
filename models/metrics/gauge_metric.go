package metrics

import "fmt"

type GaugeMetric struct {
	Name   string
	Value  int64
	Desc   string
	Labels Labels
}

func (c GaugeMetric) Key() string {
	return c.Name
}

func (c GaugeMetric) Print() string {
	return fmt.Sprintf("%s%s %d", c.Name, c.Labels.Print(), c.Value)
}

func (c GaugeMetric) Header() string {
	return fmt.Sprintf("# HELP %s %s\n# TYPE %s gauge", c.Name, c.Desc, c.Name)
}
