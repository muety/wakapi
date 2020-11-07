package models

type Filters struct {
	Project  string
	OS       string
	Language string
	Editor   string
	Machine  string
}

type FilterElement struct {
	Type uint8
	Key  string
}

func NewFiltersWith(entity uint8, key string) *Filters {
	switch entity {
	case SummaryProject:
		return &Filters{Project: key}
	case SummaryOS:
		return &Filters{OS: key}
	case SummaryLanguage:
		return &Filters{Language: key}
	case SummaryEditor:
		return &Filters{Editor: key}
	case SummaryMachine:
		return &Filters{Machine: key}
	}
	return &Filters{}
}

func (f *Filters) First() (bool, uint8, string) {
	if f.Project != "" {
		return true, SummaryProject, f.Project
	} else if f.OS != "" {
		return true, SummaryOS, f.OS
	} else if f.Language != "" {
		return true, SummaryLanguage, f.Language
	} else if f.Editor != "" {
		return true, SummaryEditor, f.Editor
	} else if f.Machine != "" {
		return true, SummaryMachine, f.Machine
	}
	return false, 0, ""
}

func (f *Filters) All() []*FilterElement {
	all := make([]*FilterElement, 0)

	if f.Project != "" {
		all = append(all, &FilterElement{Type: SummaryProject, Key: f.Project})
	}
	if f.Editor != "" {
		all = append(all, &FilterElement{Type: SummaryEditor, Key: f.Editor})
	}
	if f.Language != "" {
		all = append(all, &FilterElement{Type: SummaryLanguage, Key: f.Language})
	}
	if f.Machine != "" {
		all = append(all, &FilterElement{Type: SummaryMachine, Key: f.Machine})
	}
	if f.OS != "" {
		all = append(all, &FilterElement{Type: SummaryOS, Key: f.OS})
	}

	return all
}
