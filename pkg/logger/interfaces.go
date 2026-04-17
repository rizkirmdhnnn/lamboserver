package logger

// DebugLogger is the minimal interface other packages use for logging.
// The concrete type *Logger satisfies this interface.
// Downstream packages should accept DebugLogger in constructors to enable
// test injection of mock or no-op loggers.
type DebugLogger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Action(action string, err error)
	IsEnabled() bool
}
