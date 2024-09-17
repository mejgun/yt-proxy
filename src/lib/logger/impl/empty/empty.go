package logger

type loggerT struct {
}

func (t *loggerT) LogError(string, ...any)   {}
func (t *loggerT) LogWarning(string, ...any) {}
func (t *loggerT) LogDebug(string, ...any)   {}
func (t *loggerT) LogInfo(string, ...any)    {}

func New() (*loggerT, error) {
	return &loggerT{}, nil
}
