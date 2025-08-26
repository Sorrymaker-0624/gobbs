package logger

import (
	"go.uber.org/zap"
	"log"
)

var Log *zap.Logger

func Init() {
	var err error

	Log, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("无法初始化Zap Logger: %v", err)
	}

	zap.ReplaceGlobals(Log)
}
