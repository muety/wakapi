package models

type Filters struct {
	Project  string
	OS       string
	Language string
	Editor   string
	Machine  string
	Label    string
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
	case SummaryLabel:
		return &Filters{Label: key}
	}
	return &Filters{}
}

func (f *Filters) One() (bool, uint8, string) {
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
	} else if f.Label != "" {
		return true, SummaryLabel, f.Label
	}
	return false, 0, ""
}
