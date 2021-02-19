package main

import (
	"flag"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/model"
	"github.com/ogre0403/iscsi-target-api/pkg/rest"
	"github.com/ogre0403/iscsi-target-api/pkg/tgt"
)

var (
	commitID  = "%COMMITID%"
	buildTime = "%BUILDID%"
	c         = &model.ManagerCfg{
		CHAP: &model.CHAP{},
	}
	sc                = &model.ServerCfg{}
	port              int
	targetManagerType string
)

func parserFlags() {

	flag.Set("logtostderr", "true")
	flag.StringVar(&c.TargetConf, "target-conf-file", tgt.TARGETCONF, "target config file path")
	flag.StringVar(&c.BaseImagePath, "volume-image-path", tgt.BASEIMGPATH, "foldr to place volume image")
	flag.StringVar(&c.ThinPool, "thin-pool-name", "pool0", "thin pool name, if LVM is used")
	flag.StringVar(&targetManagerType, "manager-type", tgt.TgtdType, "target manager tool type")
	flag.IntVar(&sc.Port, "api-port", 8811, "api server port")
	flag.StringVar(&sc.Username, "api-username", "admin", "api admin name")
	flag.StringVar(&sc.Password, "api-password", "password", "api admin password")
	flag.StringVar(&c.CHAP.CHAPUser, "chap-username", "", "CHAP username for initiator")
	flag.StringVar(&c.CHAP.CHAPPassword, "chap-password", "", "CHAP password for initiator")
	flag.StringVar(&c.CHAP.CHAPUserIn, "chap-username-in", "", "CHAP username for target(s)")
	flag.StringVar(&c.CHAP.CHAPPasswordIn, "chap-password-in", "", "CHAP password for target(s)")
	flag.Parse()
}

func showVersion() {
	log.V(1).Infof("BuildTime: %s", buildTime)
	log.V(1).Infof("CommitID: %s", commitID)
}

func main() {

	parserFlags()
	showVersion()

	m, err := tgt.NewTarget(targetManagerType, c)

	if err != nil {
		log.Fatalf("initialize target manager fail: %s", err.Error())
		return
	}

	s := rest.NewAPIServer(m, sc)
	if s == nil {
		log.Fatal("Initialize server fail")
		return
	}

	err = s.RunServer()
	if err != nil {
		log.Fatalf("Start server fail: %s", err.Error())
		return
	}
}
