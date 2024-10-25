package infra

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"m3u8dl_for_web/conf"
	"os"
)

var Logger = logrus.New()

func InitLogger(conf *conf.Config) {
	logFile := &lumberjack.Logger{
		Filename: conf.Log.Path,
		MaxSize:  conf.Log.MaxSize, // MB
		//MaxBackups: conf.Log.LogNu,
		MaxAge:    conf.Log.MaxAge, // days
		Compress:  false,
		LocalTime: true,
	}

	multiWriter := io.MultiWriter(logFile, os.Stdout)

	Logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	Logger.SetReportCaller(true)
	//Logger.SetOutput(os.Stdout)
	Logger.SetOutput(multiWriter)
}
