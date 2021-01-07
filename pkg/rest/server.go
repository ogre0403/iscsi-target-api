package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/ogre0403/iscsi-target-api/pkg/tgt"
	"strconv"
)

type APIServer struct {
	router    *gin.Engine
	targetMgr *tgt.TargetManager
}

func NewAPIServer(m tgt.TargetManager) *APIServer {

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
