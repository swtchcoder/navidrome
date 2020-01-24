package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/deluan/navidrome/conf"
	"github.com/deluan/navidrome/log"
)

func Init(t *testing.T, skipOnShort bool) {
	if skipOnShort && testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	_, file, _, _ := runtime.Caller(0)
	appPath, _ := filepath.Abs(filepath.Join(filepath.Dir(file), ".."))
	confPath, _ := filepath.Abs(filepath.Join(appPath, "tests", "navidrome-test.toml"))

	os.Chdir(appPath)
	conf.LoadFromFile(confPath)

	noLog := os.Getenv("NOLOG")
	if noLog != "" {
		log.SetLevel(log.LevelError)
	}
}
