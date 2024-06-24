package logger

import (
	"log"
	"sync"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// Logger creates a new instance of zap logger
func Logger() *zap.Logger {
	once.Do(func() {
		lgr, err := zap.NewDevelopment()
		if err != nil {
			log.Println("Could not init logger", err)
			// reset once
			once = sync.Once{}
			return
		}
		logger = lgr
	})
	return logger
}

// Sugar returns zap sugar a simpler logger version
func Sugar() *zap.SugaredLogger {
	return Logger().Sugar()
}
