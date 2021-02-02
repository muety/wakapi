package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"gorm.io/gorm"
	"sort"
	"strings"
)

type migrationFunc struct {
	f    func(db *gorm.DB, cfg *config.Config) error
	name string
}

type migrationFuncs []migrationFunc

var (
	preMigrations  migrationFuncs
	postMigrations migrationFuncs
)

func registerPreMigration(f migrationFunc) {
	preMigrations = append(preMigrations, f)
}

func registerPostMigration(f migrationFunc) {
	postMigrations = append(postMigrations, f)
}

// NOTE: Currently, migrations themselves keep track
// of whether they have run, yet or not, because some
// simply run on every start.

func RunPreMigrations(db *gorm.DB, cfg *config.Config) {
	sort.Sort(preMigrations)

	for _, m := range preMigrations {
		logbuch.Info("potentially running migration '%s'", m.name)
		if err := m.f(db, cfg); err != nil {
			logbuch.Fatal("migration '%s' failed – %v", m.name, err)
		}
	}
}

func RunCustomPostMigrations(db *gorm.DB, cfg *config.Config) {
	sort.Sort(postMigrations)

	for _, m := range postMigrations {
		logbuch.Info("potentially running migration '%s'", m.name)
		if err := m.f(db, cfg); err != nil {
			logbuch.Fatal("migration '%s' failed – %v", m.name, err)
		}
	}
}

func (m migrationFuncs) Len() int {
	return len(m)
}

func (m migrationFuncs) Less(i, j int) bool {
	return strings.Compare(m[i].name, m[j].name) < 0
}

func (m migrationFuncs) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
