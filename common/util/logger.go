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
func InitLogger(path, filename string) (*logrus.Logger, error) {

	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return nil, err
		}
	}

	logFile, err := os.OpenFile(path+filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	Logger := logrus.New()
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.AddHook(&hooker{
		Writer: logFile,
		EntryLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.ErrorLevel,
		},
	})

	return Logger, nil
}
