package cfg

type ManagerCfg struct {
	BaseImagePath string
}

type VolumeCfg struct {
	Path string
	Size string
	Name string
	// todo: support in the future
	Type         string
	ThinProvsion bool
}

type LunCfg struct {
}

type TargetCfg struct {
	TargetId  string
	TargetIQN string
}
