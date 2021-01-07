package main

import (
	"flag"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/rest"
)

func main() {

	flag.Parse()
	s := rest.NewAPIServer()
	if s == nil {
		log.Error("Start server fail")
	}

	err := s.RunServer(80)
	if err != nil {
		log.Error(err.Error())
	}


}
