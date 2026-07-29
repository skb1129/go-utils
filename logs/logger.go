package logs

import (
	"fmt"

	"github.com/skb1129/go-utils/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func NewLogger() *zap.Logger {
	environment := config.GetString("environment")
	var err error
	if environment == "prod" {
		if err = zap.RegisterEncoder(environment, NewEncoder); err == nil {
			cfg := zap.NewProductionConfig()
			cfg.Encoding = environment
			cfg.OutputPaths = []string{"stdout"}
			cfg.ErrorOutputPaths = []string{"stderr"}
			cfg.EncoderConfig.TimeKey = "timestamp"
			cfg.EncoderConfig.LevelKey = "level"
			cfg.EncoderConfig.MessageKey = "message"
			cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
			cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
			logger, err = cfg.Build()
		}
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(fmt.Errorf("unable to initialize logger\n %w", err))
	}
	return logger
}

func GetLogger() *zap.Logger {
	if logger == nil {
		return NewLogger()
	}
	return logger
}
