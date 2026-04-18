package logger_test

import "github.com/stretchr/testify/mock"

// MockLogger satisfies debug.DebugLogger for use in package-internal tests
// and as a pattern for downstream packages that need a silent logger.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Error(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Action(action string, err error) {
	m.Called(action, err)
}

func (m *MockLogger) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

// NoopLogger returns a MockLogger that accepts any call without assertion.
// Use when the test doesn't care about log output, only side effects.
func NoopLogger() *MockLogger {
	l := new(MockLogger)
	l.On("Info", mock.Anything, mock.Anything).Return()
	l.On("Error", mock.Anything, mock.Anything).Return()
	l.On("Action", mock.Anything, mock.Anything).Return()
	l.On("IsEnabled").Return(false)
	return l
}
