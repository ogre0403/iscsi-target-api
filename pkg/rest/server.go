package rest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"github.com/ogre0403/go-lvm"
	"github.com/ogre0403/iscsi-target-api/pkg/manager"
	"github.com/ogre0403/iscsi-target-api/pkg/model"
	"github.com/ogre0403/iscsi-target-api/pkg/target"
	"github.com/ogre0403/iscsi-target-api/pkg/volume"
	"net/http"
	"strconv"
	"sync/atomic"
)

const (
	CreateVolumeEndpoint = "/createVol"
	AttachLunEndpoint    = "/attachLun"
	DeleteTargetEndpoint = "/deleteTar"
	DeleteVolumeEndpoint = "/deleteVol"
)

type APIServer struct {
	config    *model.ServerCfg
	router    *gin.Engine
	targetMgr *manager.TargetManager
	locker    uint32
}

func NewAPIServer(m manager.TargetManager, conf *model.ServerCfg) *APIServer {

	log.Info("initialize lvm plugin")
	err := lvm.Initialize()
	if err != nil {
		log.Errorf("initialize lvm plugin fail: %s", err.Error())
		return nil
	}

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
	authorized.POST(CreateVolumeEndpoint, server.CreateVolumeAPI)
	authorized.POST(AttachLunEndpoint, server.AttachLunAPI)
	authorized.DELETE(DeleteTargetEndpoint, server.DeleteTargetAPI)
	authorized.DELETE(DeleteVolumeEndpoint, server.DeleteVolumeAPI)
}

func (s *APIServer) CreateVolumeAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&s.locker, 0)

	var req volume.BasicVolume
	err := c.BindJSON(&req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind BasicVolume fail: %s", err.Error())
		return
	}

	if err := (*s.targetMgr).CreateVolume(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Create Volume fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func (s *APIServer) AttachLunAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&s.locker, 0)

	var req model.Lun
	err := c.BindJSON(&req)

	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind Lun fail: %s", err.Error())
		return
	}

	if err := (*s.targetMgr).AttachLun(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Attach LUN fail: %s", err.Error())
		return
	}

	if err := (*s.targetMgr).Save(); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Save tgtd config file fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func (s *APIServer) DeleteVolumeAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&s.locker, 0)

	var req volume.BasicVolume
	err := c.BindJSON(&req)

	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind BasicVolume fail: %s", err.Error())
		return
	}

	if err := (*s.targetMgr).DeleteVolume(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Delete Volume fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func (s *APIServer) DeleteTargetAPI(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		respondWithError(c, http.StatusTooManyRequests, "Another API is executing")
		return
	}
	defer atomic.StoreUint32(&s.locker, 0)

	var req target.BasicTarget
	err := c.BindJSON(&req)

	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Bind Target fail: %s", err.Error())
		return
	}

	if err := (*s.targetMgr).DeleteTarget(&req); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Delete Target fail: %s", err.Error())
		return
	}

	if err := (*s.targetMgr).Save(); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Save tgtd config file fail: %s", err.Error())
		return
	}

	responseWithOk(c)
}

func respondWithError(c *gin.Context, code int, format string, args ...interface{}) {
	log.Errorf(format, args...)
	resp := model.Response{
		Error:   true,
		Message: fmt.Sprintf(format, args...),
	}
	c.JSON(code, resp)
	c.Abort()
}

func responseWithOk(c *gin.Context) {
	c.JSON(http.StatusOK, model.Response{
		Error:   false,
		Message: "Successful",
	})
}
