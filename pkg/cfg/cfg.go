package cfg

type ManagerCfg struct {
	BaseImagePath string
	TargetConf    string
}

type VolumeCfg struct {
	Path string `json:"path"`
	Size string `json:"size"`
	Name string `json:"name"`

	// todo: support in the future
	Type          string `json:"type"`
	ThinProvision bool   `json:"thinProvision"`
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
