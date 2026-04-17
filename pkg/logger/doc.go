// Package logger provides debug logging and log file reading for LamboServer.
//
// The Logger type supports Info, Error, and Action log levels with millisecond
// timestamps. It is disabled by default and can be enabled at runtime via the
// Enable method (triggered by config.DebugMode or the frontend EnableDebug call).
//
// The Reader type reads service log files (Nginx, PHP-FPM, etc.) from disk,
// returning recent entries for display in the frontend log viewer.
package logger
