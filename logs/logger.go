package logs

import (
	"fmt"

	"github.com/skb1129/go-utils/config"
	"go.uber.org/zap"
)

var logger *zap.Logger

func NewLogger() *zap.Logger {
	environment := config.GetString("environment")

	var err error
	if environment == "prod" {
		logger, err = zap.NewProduction()
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
