package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/tgt"
	"strconv"
)

const (
	CreateVolumeEndpoint = "/createVol"
	AttachLunEndpoint    = "/attachLun"
	DeleteTargetEndpoint = "/deleteTar"
	DeleteVolumeEndpoint = "/deleteVol"
)

type APIServer struct {
	config    *cfg.ServerCfg
	router    *gin.Engine
	targetMgr *tgt.TargetManager
}

func NewAPIServer(m tgt.TargetManager, conf *cfg.ServerCfg) *APIServer {

	s := &APIServer{
		router:    gin.Default(),
		targetMgr: &m,
		config:    conf,
	}
	addRoute(s)

	return s
}

func (s *APIServer) RunServer() error {
	err := s.router.Run(":" + strconv.Itoa(s.config.Port))
	if err != nil {
		return err
	}
	return nil
}

func addRoute(server *APIServer) {
	authorized := server.router.Group("/", gin.BasicAuth(gin.Accounts{
		server.config.Username: server.config.Password,
	}))
	authorized.POST(CreateVolumeEndpoint, (*server.targetMgr).CreateVolumeAPI)
	authorized.POST(AttachLunEndpoint, (*server.targetMgr).AttachLunAPI)
	authorized.DELETE(DeleteTargetEndpoint, (*server.targetMgr).DeleteTargetAPI)
	authorized.DELETE(DeleteVolumeEndpoint, (*server.targetMgr).DeleteVolumeAPI)
}
