package models

import "sort"

type Heartbeats []*Heartbeat

func (h Heartbeats) Len() int {
	return len(h)
}

func (h Heartbeats) Less(i, j int) bool {
	return h[i].Time.T().Before(h[j].Time.T())
}

func (h Heartbeats) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *Heartbeats) Sorted() *Heartbeats {
	sort.Sort(h)
	return h
}

func (h *Heartbeats) First() *Heartbeat {
	// assumes slice to be sorted
	if h.Len() == 0 {
		return nil
	}
	return (*h)[0]
}

func (h *Heartbeats) Last() *Heartbeat {
	// assumes slice to be sorted
	if h.Len() == 0 {
		return nil
	}
	return (*h)[h.Len()-1]
}
