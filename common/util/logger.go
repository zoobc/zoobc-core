package util

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type (
	hooker struct {
		Writer      *os.File
		EntryLevels []logrus.Level
	}
)

func (h hooker) Fire(entry *logrus.Entry) error {
	line, err := entry.String()

	if err != nil {
		return fmt.Errorf("failed on entry, %s", err.Error())
	}

	_, err = h.Writer.Write([]byte(line))
	if err != nil {
		return fmt.Errorf("failed on write entry, %s", err.Error())
	}
	return nil
}

func (h hooker) Levels() []logrus.Level {
	return h.EntryLevels
}

/*
InitLogger is function that should be implemeneted with interceptor. That can centralized the log action.
`[]logrus.Level` can inject dynamically switch on development or production mode
*/
func InitLogger(path, filename string, levels []string) (*logrus.Logger, error) {
	var (
		logLevels   []logrus.Level
		lowestLevel logrus.Level
		logger      *logrus.Logger
		err         error
		logFile     *os.File
	)
	_, err = os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return nil, err
		}
	}

	logFile, err = os.OpenFile(path+filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	logger = logrus.New()
	for _, v := range levels {
		switch v {
		case "info":
			logLevels = append(logLevels, logrus.InfoLevel)
		case "warn":
			logLevels = append(logLevels, logrus.WarnLevel)
		case "error":
			logLevels = append(logLevels, logrus.ErrorLevel)
		case "fatal":
			logLevels = append(logLevels, logrus.FatalLevel)
		case "panic":
			logLevels = append(logLevels, logrus.PanicLevel)
		case "debug":
			logLevels = append(logLevels, logrus.DebugLevel)
		}
		// lowestLevel will based on the list log level will use
		if lowestLevel < logLevels[len(logLevels)-1] {
			lowestLevel = logLevels[len(logLevels)-1]
		}
	}
	if len(logLevels) < 1 {
		logLevels = append(
			logLevels,
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		)
	}
	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.AddHook(&hooker{
		Writer:      logFile,
		EntryLevels: logLevels,
	})
	// lowestLevel use to set lowest level will fire
	logger.SetLevel(lowestLevel)

	return logger, nil
}
