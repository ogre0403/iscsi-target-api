package main

import (
	"flag"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/tgt"
)

func main() {

	flag.Parse()
	m, err := tgt.NewTarget("tgtd", &cfg.ManagerCfg{
		BaseImagePath: "/var/lib/iscsi/",
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	err = m.CreateVolume(&cfg.VolumeCfg{
		Name: "test.img",
		Size: "100m",
	})
	if err != nil {
		log.Fatal(err.Error())
	}

}
