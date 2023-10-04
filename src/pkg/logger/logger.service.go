package logger

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warning
	Error
	Fatal
)

func LogLevelFromString(level string) LogLevel {
	switch level {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warning":
		return Warning
	case "error":
		return Error
	case "fatal":
		return Fatal
	default:
		return Info
	}
}

type Config struct {
	Level LogLevel
}

type Logger interface {
	Info(format string, v ...any)
	Error(format string, v ...any)
	Warning(format string, v ...any)
	Debug(format string, v ...any)
	Fatal(format string, v ...any)
}
