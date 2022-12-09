package logger

import (
	"os"

	"go.uber.org/zap"
)

var err error

type logger struct {
	*zap.Logger
}

func (l *logger) init() error {
	if os.Getenv("MODE") == "development" {
		l.Logger, err = zap.NewDevelopment()
	} else {
		l.Logger, err = zap.NewProduction()
	}

	return err
}

// New takes in a package to initlialize the new logger in.
func New(pkg string) *logger {
	Log := &logger{}
	err = Log.init()
	if err != nil {
		panic(err)
	}

	Log.Logger = Log.Logger.With(
		zap.String("package", pkg),
	)

	return Log
}
