package infra

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger initializes logging. Provide concrete log configuration instead of importing conf to avoid import cycles.
func InitLogger(logPath string, maxSize int, maxAge int) {
	logFile := &lumberjack.Logger{
		Filename:  logPath,
		MaxSize:   maxSize, // MB
		MaxAge:    maxAge,  // days
		Compress:  false,
		LocalTime: true,
	}

	originalStdout, _, err := interceptStdio()
	if err != nil {
		panic(err)
	}
	multiWriter := io.MultiWriter(logFile, originalStdout)

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logrus.SetLevel(logrus.DebugLevel)

	logrus.SetReportCaller(true)
	// Logger.SetOutput(os.Stdout)
	logrus.SetOutput(multiWriter)
}

// 能拦截golang的stdio输出,并不能拦截CGO调用动态链接库的输出
func interceptStdio() (*os.File, *os.File, error) {
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	writer := logrus.StandardLogger().Writer()
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	os.Stdout = w

	go func() {
		for {
			io.Copy(writer, r)
		}
	}()

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	os.Stderr = stderrWriter

	go func() {
		for {
			io.Copy(writer, stderrReader)
		}
	}()

	return originalStdout, originalStderr, nil
}
