package infra

import (
	"m3u8dl_for_web/conf"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger(conf conf.Config) {
	// log_nu := conf.Log.LogNu

	// logFile := &lumberjack.Logger{
	// 	Filename:   conf.Log.Path,
	// 	MaxSize:    10, // MB
	// 	MaxBackups: log_nu,
	// 	MaxAge:     28, // days
	// 	Compress:   true,
	// 	LocalTime:  true,
	// }

	Logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	Logger.SetReportCaller(true)
	Logger.SetOutput(os.Stdout)
}
