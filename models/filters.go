package models

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/mitchellh/hashstructure/v2"
)

type Filters struct {
	Project  OrFilter
	OS       OrFilter
	Language OrFilter
	Editor   OrFilter
	Machine  OrFilter
	Label    OrFilter
	Branch   OrFilter
	Entity   OrFilter
}

type OrFilter []string

func (f OrFilter) Exists() bool {
	return len(f) > 0 && f[0] != ""
}

func (f OrFilter) MatchAny(search string) bool {
	for _, s := range f {
		if s == search || (s == "-" && search == "") {
			return true
		}
	}
	return false
}

type FilterElement struct {
	entity uint8
	filter OrFilter
}

func NewFiltersWith(entity uint8, key string) *Filters {
	return NewFilterWithMultiple(entity, []string{key})
}

func NewFilterWithMultiple(entity uint8, keys []string) *Filters {
	filters := &Filters{}
	return filters.WithMultiple(entity, keys)
}

func (f *Filters) With(entity uint8, key string) *Filters {
	return f.WithMultiple(entity, []string{key})
}

func (f *Filters) WithMultiple(entity uint8, keys []string) *Filters {
	switch entity {
	case SummaryProject:
		f.Project = append(f.Project, keys...)
	case SummaryOS:
		f.OS = append(f.OS, keys...)
	case SummaryLanguage:
		f.Language = append(f.Language, keys...)
	case SummaryEditor:
		f.Editor = append(f.Editor, keys...)
	case SummaryMachine:
		f.Machine = append(f.Machine, keys...)
	case SummaryLabel:
		f.Label = append(f.Label, keys...)
	case SummaryBranch:
		f.Branch = append(f.Branch, keys...)
	case SummaryEntity:
		f.Entity = append(f.Entity, keys...)
	}
	return f
}

func (f *Filters) One() (bool, uint8, OrFilter) {
	if f.Project != nil && f.Project.Exists() {
		return true, SummaryProject, f.Project
	} else if f.OS != nil && f.OS.Exists() {
		return true, SummaryOS, f.OS
	} else if f.Language != nil && f.Language.Exists() {
		return true, SummaryLanguage, f.Language
	} else if f.Editor != nil && f.Editor.Exists() {
		return true, SummaryEditor, f.Editor
	} else if f.Machine != nil && f.Machine.Exists() {
		return true, SummaryMachine, f.Machine
	} else if f.Label != nil && f.Label.Exists() {
		return true, SummaryLabel, f.Label
	} else if f.Branch != nil && f.Branch.Exists() {
		return true, SummaryBranch, f.Branch
	} else if f.Entity != nil && f.Entity.Exists() {
		return true, SummaryEntity, f.Entity
	}
	return false, 0, OrFilter{}
}

func (f *Filters) OneOrEmpty() FilterElement {
	if ok, t, of := f.One(); ok {
		return FilterElement{entity: t, filter: of}
	}
	return FilterElement{entity: SummaryUnknown, filter: []string{}}
}

func (f *Filters) IsEmpty() bool {
	nonEmpty, _, _ := f.One()
	return !nonEmpty
}

func (f *Filters) Count() int {
	var count int
	for i := SummaryProject; i <= SummaryEntity; i++ {
		count += f.CountByEntity(i)
	}
	return count
}

func (f *Filters) CountByEntity(entity uint8) int {
	return len(*f.ResolveEntity(entity))
}

func (f *Filters) EntityCount() int {
	var count int
	for i := SummaryProject; i <= SummaryEntity; i++ {
		if c := f.CountByEntity(i); c > 0 {
			count++
		}
	}
	return count
}

func (f *Filters) ResolveEntity(entityId uint8) *OrFilter {
	switch entityId {
	case SummaryProject:
		return &f.Project
	case SummaryLanguage:
		return &f.Language
	case SummaryEditor:
		return &f.Editor
	case SummaryOS:
		return &f.OS
	case SummaryMachine:
		return &f.Machine
	case SummaryLabel:
		return &f.Label
	case SummaryBranch:
		return &f.Branch
	case SummaryEntity:
		return &f.Entity
	default:
		return &OrFilter{}
	}
}

func (f *Filters) Hash() string {
	hash, err := hashstructure.Hash(f, hashstructure.FormatV2, nil)
	if err != nil {
		logbuch.Error("CRITICAL ERROR: failed to hash struct - %v", err)
	}
	return fmt.Sprintf("%x", hash) // "uint64 values with high bit set are not supported"
}

func (f *Filters) Match(h *Heartbeat) bool {
	return (f.Project == nil || f.Project.MatchAny(h.Project)) &&
		(f.OS == nil || f.OS.MatchAny(h.OperatingSystem)) &&
		(f.Language == nil || f.Language.MatchAny(h.Language)) &&
		(f.Editor == nil || f.Editor.MatchAny(h.Editor)) &&
		(f.Machine == nil || f.Machine.MatchAny(h.Machine))
}

// WithAliases adds OR-conditions for every alias of a filter key as additional filter keys
func (f *Filters) WithAliases(resolve AliasReverseResolver) *Filters {
	if f.Project != nil {
		updated := OrFilter(make([]string, 0, len(f.Project)))
		for _, e := range f.Project {
			updated = append(updated, e)
			updated = append(updated, resolve(SummaryProject, e)...)
		}
		f.Project = updated
	}
	if f.OS != nil {
		updated := OrFilter(make([]string, 0, len(f.OS)))
		for _, e := range f.OS {
			updated = append(updated, e)
			updated = append(updated, resolve(SummaryOS, e)...)
		}
		f.OS = updated
	}
	if f.Language != nil {
		updated := OrFilter(make([]string, 0, len(f.Language)))
		for _, e := range f.Language {
			updated = append(updated, e)
			updated = append(updated, resolve(SummaryLanguage, e)...)
		}
		f.Language = updated
	}
	if f.Editor != nil {
		updated := OrFilter(make([]string, 0, len(f.Editor)))
		for _, e := range f.Editor {
			updated = append(updated, e)
			updated = append(updated, resolve(SummaryEditor, e)...)
		}
		f.Editor = updated
	}
	if f.Machine != nil {
		updated := OrFilter(make([]string, 0, len(f.Machine)))
		for _, e := range f.Machine {
			updated = append(updated, e)
			updated = append(updated, resolve(SummaryMachine, e)...)
		}
		f.Machine = updated
	}
	if f.Branch != nil {
		updated := OrFilter(make([]string, 0, len(f.Branch)))
		for _, e := range f.Branch {
			updated = append(updated, e)
			updated = append(updated, resolve(SummaryBranch, e)...)
		}
		f.Branch = updated
	}
	// no aliases for entites / files
	return f
}

func (f *Filters) WithProjectLabels(resolve ProjectLabelReverseResolver) *Filters {
	if f.Label == nil || !f.Label.Exists() {
		return f
	}
	for _, l := range f.Label {
		f.WithMultiple(SummaryProject, resolve(l))
	}
	return f
}

func (f *Filters) IsProjectDetails() bool {
	return f != nil && f.Project != nil && f.Project.Exists()
}
