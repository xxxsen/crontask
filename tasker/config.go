package tasker

type config struct {
	enableUserCred bool
	uid            uint64
	gid            uint64
	workdir        string
	executor       string
	params         []string
	expression     string
}

type Option func(c *config)

func WithUserGroup(uid, gid uint64) Option {
	return func(c *config) {
		c.enableUserCred = true
		c.uid, c.gid = uid, gid
	}
}

func WithProgram(exec string, args []string) Option {
	return func(c *config) {
		c.executor = exec
		c.params = args
	}
}

func WithWorkDir(dir string) Option {
	return func(c *config) {
		c.workdir = dir
	}
}

func WithCronExpression(exp string) Option {
	return func(c *config) {
		c.expression = exp
	}
}