package logging

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logger = logrus.New()

func Setup() {
	Logger := logrus.New()
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.InfoLevel)
}
