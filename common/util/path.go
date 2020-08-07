package util

import (
	"os"
	"path"
	"strings"
)

func GetRootPath() string {
	wd, _ := os.Getwd()
	if strings.Contains(wd, "zoobc-core/") {
		return path.Join(wd, "../")
	}
	return wd
}
