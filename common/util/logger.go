// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package util

import (
	"fmt"
	"os"
	"path/filepath"

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
InitLogger is function that should be implemented with interceptor. That can centralized the log action.
`[]logrus.Level` can inject dynamically switch on development or production mode
*/
func InitLogger(path, filename string, levels []string, logOnCLI bool) (*logrus.Logger, error) {
	var (
		logLevels   []logrus.Level
		lowestLevel logrus.Level
		logger      *logrus.Logger
		err         error
		logFile     *os.File
	)
	_, err = os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		if e := os.Mkdir(path, os.ModePerm); e != nil {
			return nil, e
		}
	}
	logFile, err = os.OpenFile(filepath.Join(path, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil, err
	}

	logger = logrus.New()
	for _, v := range levels {
		switch v {
		case "debug":
			logLevels = append(logLevels, logrus.DebugLevel)
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
	if logOnCLI {
		logger.AddHook(&hooker{
			Writer:      logFile,
			EntryLevels: logLevels,
		})
	} else {
		// only record log on file
		logger.SetOutput(logFile)
	}
	logger.ExitFunc = func(i int) {
		if logFile == nil {
			return
		}
		_ = logFile.Close()
	}
	// lowestLevel use to set lowest level will fire
	logger.SetLevel(lowestLevel)

	return logger, nil
}
