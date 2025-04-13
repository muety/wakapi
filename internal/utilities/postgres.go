package utilities

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	conf "github.com/muety/wakapi/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresError is a custom error struct for marshalling Postgres errors to JSON.
type PostgresError struct {
	Code           string `json:"code"`
	HttpStatusCode int    `json:"-"`
	Message        string `json:"message"`
	Hint           string `json:"hint,omitempty"`
	Detail         string `json:"detail,omitempty"`
}

// NewPostgresError returns a new PostgresError if the error was from a publicly
// accessible Postgres error.
func NewPostgresError(err error) *PostgresError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && isPubliclyAccessiblePostgresError(pgErr.Code) {
		return &PostgresError{
			Code:           pgErr.Code,
			HttpStatusCode: getHttpStatusCodeFromPostgresErrorCode(pgErr.Code),
			Message:        pgErr.Message,
			Detail:         pgErr.Detail,
			Hint:           pgErr.Hint,
		}
	}

	return nil
}
func (pg *PostgresError) IsUniqueConstraintViolated() bool {
	// See https://www.postgresql.org/docs/current/errcodes-appendix.html for list of error codes
	return pg.Code == "23505"
}

// isPubliclyAccessiblePostgresError checks if the Postgres error should be
// made accessible.
func isPubliclyAccessiblePostgresError(code string) bool {
	if len(code) != 5 {
		return false
	}

	// default response
	return getHttpStatusCodeFromPostgresErrorCode(code) != 0
}

// getHttpStatusCodeFromPostgresErrorCode maps a Postgres error code to a HTTP
// status code. Returns 0 if the code doesn't map to a given postgres error code.
func getHttpStatusCodeFromPostgresErrorCode(code string) int {
	if code == pgerrcode.RaiseException ||
		code == pgerrcode.IntegrityConstraintViolation ||
		code == pgerrcode.RestrictViolation ||
		code == pgerrcode.NotNullViolation ||
		code == pgerrcode.ForeignKeyViolation ||
		code == pgerrcode.UniqueViolation ||
		code == pgerrcode.CheckViolation ||
		code == pgerrcode.ExclusionViolation {
		return 500
	}

	// Use custom HTTP status code if Postgres error was triggered with `PTXXX`
	// code. This is consistent with PostgREST's behaviour as well.
	if strings.HasPrefix(code, "PT") {
		if httpStatusCode, err := strconv.ParseInt(code[2:], 10, 0); err == nil {
			return int(httpStatusCode)
		}
	}

	return 0
}

func GetGormLogger(sql string) logger.Interface {
	sqlLog := logrus.WithField("component", "sql")

	// Determine SQL logging behavior based on the configuration
	var gormLogLevel logger.LogLevel
	switch sql {
	case "all":
		gormLogLevel = logger.Info
		sqlLog.Debug("SQL logging enabled for all statements and arguments")
	case "statement":
		gormLogLevel = logger.Info
		sqlLog.Debug("SQL logging enabled for statements only")
	case "none":
		gormLogLevel = logger.Silent
		sqlLog.Debug("SQL logging disabled")
	default:
		gormLogLevel = logger.Silent
		sqlLog.Warn("Unknown SQL logging level, defaulting to 'none'")
	}

	// Create and return the GORM logger
	return logger.New(
		logrus.StandardLogger(),
		logger.Config{
			SlowThreshold:             time.Minute, // Log queries slower than this threshold
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,  // Ignore ErrRecordNotFound errors
			Colorful:                  false, // Disable colorful output
		},
	)
}

func InitDB(config *conf.Config) (*gorm.DB, *sql.DB, error) {
	// Set up GORM
	gormLogger := GetGormLogger(config.Logging.SQL)

	// Connect to database
	var err error
	slog.Info("starting with database", "dialect", config.Db.Dialect)
	db, err := gorm.Open(config.Db.GetDialector(), &gorm.Config{Logger: gormLogger}, conf.GetWakapiDBOpts(&config.Db))
	if err != nil {
		// Use NewPostgresError to check if it's a Postgres-specific error
		if pgErr := NewPostgresError(err); pgErr != nil {
			slog.Error("Postgres error occurred", "code", pgErr.Code, "message", pgErr.Message, "hint", pgErr.Hint, "detail", pgErr.Detail)
			return nil, nil, fmt.Errorf("postgres error: %s (code: %s)", pgErr.Message, pgErr.Code)
		}

		// If it's not a Postgres-specific error, return the generic error
		conf.Log().Fatal("could not connect to database", "error", err)
		return nil, nil, err
	}

	if config.IsDev() {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		// Use NewPostgresError to check if it's a Postgres-specific error
		if pgErr := NewPostgresError(err); pgErr != nil {
			slog.Error("Postgres error occurred", "code", pgErr.Code, "message", pgErr.Message, "hint", pgErr.Hint, "detail", pgErr.Detail)
			return nil, nil, fmt.Errorf("postgres error: %s (code: %s)", pgErr.Message, pgErr.Code)
		}

		// If it's not a Postgres-specific error, return the generic error
		conf.Log().Fatal("could not connect to database", "error", err)
		return nil, nil, err
	}

	sqlDB.SetMaxIdleConns(int(config.Db.MaxConn))
	sqlDB.SetMaxOpenConns(int(config.Db.MaxConn))

	return db, sqlDB, nil
}
