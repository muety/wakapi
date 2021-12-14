package models

import "sort"

type Durations []*Duration

func (d Durations) Len() int {
	return len(d)
}

func (d Durations) Less(i, j int) bool {
	return d[i].Time.T().Before(d[j].Time.T())
}

func (d Durations) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d *Durations) Sorted() *Durations {
	sort.Sort(d)
	return d
}

func (d *Durations) First() *Duration {
	// assumes slice to be sorted
	if d.Len() == 0 {
		return nil
	}
	return (*d)[0]
}

func (d *Durations) Last() *Duration {
	// assumes slice to be sorted
	if d.Len() == 0 {
		return nil
	}
	return (*d)[d.Len()-1]
}
