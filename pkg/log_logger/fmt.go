package log_logger

import (
	"encoding/json"
	"fmt"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"log"
)

type fmtLogger struct {
	logger *log.Logger
	config *logger.Config
}

func New(l *log.Logger, config *logger.Config) logger.Logger {
	return &fmtLogger{l, config}
}

func (l *fmtLogger) Info(format string, v ...any) {
	if l.config.Level > logger.Info {
		return
	}
	l.print("INFO: "+format, l.prettyAll(v...)...)
}

func (l *fmtLogger) Error(format string, v ...any) {
	if l.config.Level > logger.Error {
		return
	}
	l.print("ERROR: "+format, l.prettyAll(v...)...)
}

func (l *fmtLogger) Warning(format string, v ...any) {
	if l.config.Level > logger.Warning {
		return
	}
	l.print("WARNING: "+format, l.prettyAll(v...)...)
}

func (l *fmtLogger) Debug(format string, v ...any) {
	if l.config.Level > logger.Debug {
		return
	}
	l.print("DEBUG: "+format, l.prettyAll(v...)...)
}

func (l *fmtLogger) Fatal(format string, v ...any) {
	if l.config.Level > logger.Fatal {
		return
	}
	l.logger.Fatalf("FATAL: "+format, l.prettyAll(v...)...)
}

func (l *fmtLogger) print(format string, v ...any) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in logger")
			fmt.Println("ORIGIN: "+format, v)
			fmt.Println("Stack Trace: ", r)
		}
	}()
	l.logger.Printf(format, v...)
}

func (l *fmtLogger) prettyAll(v ...any) []any {
	var p []any
	for _, data := range v {
		p = append(p, l.pretty(data))
	}
	return p
}

func (l *fmtLogger) pretty(data any) any {
	var p []byte
	p, err := json.Marshal(data)
	if err != nil {
		return data
	}
	return string(p)
}
