package cfg

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
	TargetIQN string     `json:"targetIQN"`
	Volume    *VolumeCfg `json:"volume"`
	AclIpList []string   `json:"aclList"`
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
