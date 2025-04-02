package observability

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bombsimon/logrusr/v3"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm/logger"
)

const (
	LOG_SQL_ALL       = "all"
	LOG_SQL_NONE      = "none"
	LOG_SQL_STATEMENT = "statement"
)

var (
	loggingOnce sync.Once
)

type CustomFormatter struct {
	logrus.JSONFormatter
}

func NewCustomFormatter() *CustomFormatter {
	return &CustomFormatter{
		JSONFormatter: logrus.JSONFormatter{
			DisableTimestamp: false,
			TimestampFormat:  time.RFC3339,
		},
	}
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// logrus doesn't support formatting the time in UTC so we need to use a custom formatter
	entry.Time = entry.Time.UTC()
	return f.JSONFormatter.Format(entry)
}

func ConfigureLogging(config *conf.LoggingConfig) error {
	var err error

	loggingOnce.Do(func() {
		formatter := NewCustomFormatter()
		logrus.SetFormatter(formatter)

		// Ensure the log file is written to the "log" folder in the root of the project
		if config.File != "" {
			// Determine the root directory of the project
			rootDir, errRoot := os.Getwd() // Get the current working directory
			fmt.Println("rootDir", rootDir)
			if errRoot != nil {
				err = fmt.Errorf("failed to determine root directory: %w", errRoot)
				return
			}

			// Create the "log" folder if it doesn't exist
			logDir := rootDir + "/log"
			if _, errDir := os.Stat(logDir); os.IsNotExist(errDir) {
				if errMkDir := os.MkdirAll(logDir, 0755); errMkDir != nil {
					err = fmt.Errorf("failed to create log directory: %w", errMkDir)
					return
				}
			}

			// Set the log file path
			logFilePath := logDir + "/" + config.File
			f, errOpen := os.OpenFile(logFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660) //#nosec G302 -- Log files should be rw-rw-r--
			if errOpen != nil {
				err = fmt.Errorf("failed to open log file: %w", errOpen)
				return
			}
			logrus.SetOutput(f)
			logrus.Infof("Set output file to %s", logFilePath)
		}

		// Set log level
		if config.Level != "" {
			level, errParse := logrus.ParseLevel(config.Level)
			if errParse != nil {
				err = fmt.Errorf("failed to parse log level: %w", errParse)
				return
			}
			logrus.SetLevel(level)
			logrus.Debug("Set log level to: " + logrus.GetLevel().String())
		}

		// Add custom fields to logs
		f := logrus.Fields{}
		for k, v := range config.Fields {
			f[k] = v
		}
		logrus.WithFields(f)

		// Configure GORM logger
		setGormLogger(config.SQL)

		// Set OpenTelemetry logger
		otel.SetLogger(logrusr.New(logrus.StandardLogger().WithField("component", "otel")))
	})

	return err
}

func setGormLogger(sql string) {
	sqlLog := logrus.WithField("component", "sql")

	shouldLogSQL := sql == LOG_SQL_STATEMENT || sql == LOG_SQL_ALL
	shouldLogSQLArgs := sql == LOG_SQL_ALL

	gormLogger := utilities.GetGormLogger(sql)

	if shouldLogSQL {
		// Customize GORM logger to include SQL arguments if needed
		gormLogger = gormLogger.LogMode(logger.Info)
		if shouldLogSQLArgs {
			sqlLog.Debug("SQL arguments logging enabled")
		}
	}

	// Set the GORM logger globally
	logger.Default = gormLogger
}
