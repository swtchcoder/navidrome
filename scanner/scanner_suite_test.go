package scanner

import (
	"testing"

	"github.com/astaxie/beego/orm"
	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/db"

	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestScanner(t *testing.T) {
	tests.Init(t, true)
	conf.Server.DbPath = "file::memory:?cache=shared"
	_ = orm.RegisterDataBase("default", db.Driver, conf.Server.DbPath)
	db.EnsureLatestVersion()
	log.SetLevel(log.LevelCritical)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scanner Suite")
}
