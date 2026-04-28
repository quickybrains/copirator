package zerolog

import (
	"github.com/rs/zerolog"

	"github.com/quickybrains/copirator/internal/log"
)

type noopLogger struct {
	lg *zerolog.Logger
}

func NewNoop(conf *Config) *noopLogger {
	lg := zerolog.Nop()
	return &noopLogger{
		lg: &lg,
	}
}

func (zlg *noopLogger) Info() log.LoggerField {
	logField := zlg.lg.Info()

	return &loggerField{
		field: logField,
	}
}

func (zlg *noopLogger) Trace() log.LoggerField {
	logField := zlg.lg.Trace()

	return &loggerField{
		field: logField,
	}
}

func (zlg *noopLogger) Warn() log.LoggerField {
	logField := zlg.lg.Warn()

	return &loggerField{
		field: logField,
	}
}

func (zlg *noopLogger) Debug() log.LoggerField {
	logField := zlg.lg.Debug()

	return &loggerField{
		field: logField,
	}
}

func (zlg *noopLogger) Error() log.LoggerField {
	logField := zlg.lg.Error()

	return &loggerField{
		field: logField,
	}
}

func (zlg *noopLogger) Err(err error) log.LoggerField {
	logField := zlg.lg.Err(err)

	return &loggerField{
		field: logField,
	}
}

func (zlg *noopLogger) With(key, value string) log.Logger {
	newLogger := zlg.lg.With().Str(key, value).Logger()
	return &logger{
		lg: &newLogger,
	}
}
