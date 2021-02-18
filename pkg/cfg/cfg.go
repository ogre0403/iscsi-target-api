package cfg

import "github.com/ogre0403/iscsi-target-api/pkg/volume"

// CHAP user must be create in advance
type CHAP struct {
	CHAPUser       string
	CHAPPassword   string
	CHAPUserIn     string
	CHAPPasswordIn string
}

type ManagerCfg struct {
	CHAP           *CHAP
	CHAPPasswordIn string
	BaseImagePath  string
	TargetConf     string
	ThinPool       string
}

type LunCfg struct {
	TargetIQN  string              `json:"targetIQN"`
	Volume     *volume.BasicVolume `json:"volume"`
	AclIpList  []string            `json:"aclList"`
	EnableChap bool                `json:"enableCHAP"`
}

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type ServerCfg struct {
	Port     int
	Username string
	Password string
}
