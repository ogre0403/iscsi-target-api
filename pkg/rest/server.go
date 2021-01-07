package rest

import (
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/tgt"
	"strconv"
)

type APIServer struct {
	router    *gin.Engine
	targetMgr *tgt.TargetManager
}

func NewAPIServer() *APIServer {

	m, err := tgt.NewTarget("tgtd", &cfg.ManagerCfg{
		BaseImagePath: tgt.BASEIMGPATH,
		TargetConf:    tgt.TARGETCONF,
	})

	if err != nil {
		log.Fatalf("initialize target manager fail: %s", err.Error())
		return nil
	}

	s := &APIServer{
		router:    gin.Default(),
		targetMgr: &m,
	}
	addRoute(s.router, m)

	return s
}

func (s *APIServer) RunServer(port int) error {
	err := s.router.Run(":" + strconv.Itoa(port))
	if err != nil {
		return err
	}
	return nil
}

func addRoute(r *gin.Engine, m tgt.TargetManager) {
	r.POST("/createVol", m.CreateVolumeAPI)
	r.POST("/attachLun", m.AttachLunAPI)
	r.DELETE("/deleteTar", m.DeleteTargetAPI)
	r.DELETE("/deleteVol", m.DeleteVolumeAPI)
}
