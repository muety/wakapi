package models

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	"github.com/gohugoio/hashstructure"
	"log/slog"
)

type Filters struct {
	Project            OrFilter
	OS                 OrFilter
	Language           OrFilter
	Editor             OrFilter
	Machine            OrFilter
	Label              OrFilter
	Branch             OrFilter
	Entity             OrFilter
	Category           OrFilter
	SelectFilteredOnly bool // flag indicating to drop all Entity types from a summary except the single one filtered by
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
	Entity uint8
	Filter OrFilter
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

func (f *Filters) WithSelectFilteredOnly() *Filters {
	// use with caution: setting this usually only makes sense when interested only in the entity-specific part of a summary
	// e.g. when only wanting to retrieve the total time coded in a certain language, while disregarding projects, etc.
	if f.CountDistinctTypes() <= 1 {
		f.SelectFilteredOnly = true
	}
	return f
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
	case SummaryCategory:
		f.Category = append(f.Category, keys...)
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
	} else if f.Category != nil && f.Category.Exists() {
		return true, SummaryCategory, f.Category
	}
	return false, 0, OrFilter{}
}

func (f *Filters) OneOrEmpty() FilterElement {
	if ok, t, of := f.One(); ok {
		return FilterElement{Entity: t, Filter: of}
	}
	return FilterElement{Entity: SummaryUnknown, Filter: []string{}}
}

func (f *Filters) IsEmpty() bool {
	nonEmpty, _, _ := f.One()
	return !nonEmpty
}

func (f *Filters) Count() int {
	var count int
	for i := SummaryProject; i <= SummaryEntity; i++ {
		count += f.CountByType(i)
	}
	return count
}

func (f *Filters) CountDistinctTypes() int {
	var count int
	for i := SummaryProject; i <= SummaryEntity; i++ {
		if f.CountByType(i) > 0 {
			count += f.CountByType(i)
		}
	}
	return count
}

func (f *Filters) CountByType(entity uint8) int {
	return len(*f.ResolveType(entity))
}

func (f *Filters) EntityCount() int {
	var count int
	for i := SummaryProject; i <= SummaryEntity; i++ {
		if c := f.CountByType(i); c > 0 {
			count++
		}
	}
	return count
}

func (f *Filters) ResolveType(entityId uint8) *OrFilter {
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
	case SummaryCategory:
		return &f.Category
	default:
		return &OrFilter{}
	}
}

func (f *Filters) Hash() string {
	hash, err := hashstructure.Hash(f, &hashstructure.HashOptions{Hasher: xxhash.New()})
	if err != nil {
		slog.Error("CRITICAL ERROR: failed to hash struct", "error", err)
	}
	return fmt.Sprintf("%x", hash) // "uint64 values with high bit set are not supported"
}

func (f *Filters) MatchHeartbeat(h *Heartbeat) bool {
	return (f.Project == nil || f.Project.MatchAny(h.Project)) &&
		(f.OS == nil || f.OS.MatchAny(h.OperatingSystem)) &&
		(f.Language == nil || f.Language.MatchAny(h.Language)) &&
		(f.Editor == nil || f.Editor.MatchAny(h.Editor)) &&
		(f.Machine == nil || f.Machine.MatchAny(h.Machine)) &&
		(f.Category == nil || f.Machine.MatchAny(h.Category))
}

func (f *Filters) MatchDuration(d *Duration) bool {
	return (f.Project == nil || f.Project.MatchAny(d.Project)) &&
		(f.OS == nil || f.OS.MatchAny(d.OperatingSystem)) &&
		(f.Language == nil || f.Language.MatchAny(d.Language)) &&
		(f.Editor == nil || f.Editor.MatchAny(d.Editor)) &&
		(f.Machine == nil || f.Machine.MatchAny(d.Machine)) &&
		(f.Category == nil || f.Category.MatchAny(d.Category))
}

// WithAliases adds OR-conditions for every alias of a Filter key as additional Filter keys
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
	if f.Category != nil {
		updated := OrFilter(make([]string, 0, len(f.Category)))
		for _, e := range f.Category {
			updated = append(updated, e)
			updated = append(updated, resolve(SummaryCategory, e)...)
		}
		f.Category = updated
	}
	// no aliases for entities / files
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
