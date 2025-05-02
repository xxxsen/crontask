package tasker

type TaskConfig struct {
	Remark  string
	WorkDir string
	Cmd     string
	Args    []string
}

type JobConfig struct {
	Name         string
	SubTasks     []*TaskConfig
	Expr         string
	RunWhenStart bool
}
