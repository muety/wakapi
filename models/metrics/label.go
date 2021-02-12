package metrics

import (
	"fmt"
	"strings"
)

type Labels []Label

type Label struct {
	Key   string
	Value string
}

func (l Labels) Print() string {
	printedLabels := make([]string, len(l))
	for i, e := range l {
		printedLabels[i] = e.Print()
	}
	if len(l) == 0 {
		return ""
	}
	return fmt.Sprintf("{%s}", strings.Join(printedLabels, ","))
}

func (l Label) Print() string {
	return fmt.Sprintf("%s=\"%s\"", l.Key, l.Value)
}
