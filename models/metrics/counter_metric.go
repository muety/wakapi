package metrics

import "fmt"

type CounterMetric struct {
	Name   string
	Value  int
	Desc   string
	Labels Labels
}

func (c CounterMetric) Key() string {
	return c.Name
}

func (c CounterMetric) Print() string {
	return fmt.Sprintf("%s%s %d", c.Name, c.Labels.Print(), c.Value)
}

func (c CounterMetric) Header() string {
	return fmt.Sprintf("# HELP %s %s\n# TYPE %s counter", c.Name, c.Desc, c.Name)
}
