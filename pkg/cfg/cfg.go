package cfg

type ManagerCfg struct {
	BaseImagePath string
	TargetConf    string
}

type VolumeCfg struct {
	Group string `json:"group"`
	Size  uint64 `json:"size"`
	Unit  string `json:"unit"`
	Name  string `json:"name"`
	Type  string `json:"type"`

	// todo: support in the future
	ThinProvision bool `json:"thinProvision"`
}

type LunCfg struct {
	TargetIQN string     `json:"targetIQN"`
	Volume    *VolumeCfg `json:"volume"`
}

type TargetCfg struct {
	TargetId  string `json:"-"`
	TargetIQN string `json:"targetIQN"`
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
