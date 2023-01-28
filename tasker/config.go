package tasker

type config struct {
	enableUserCred bool
	uid            uint64
	gid            uint64
	workdir        string
	executor       string
	params         []string
	expression     string
	runWhenStart   bool
	redirectStdOut string
	redirectStdErr string
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

func WithRunWhenStart(v bool) Option {
	return func(c *config) {
		c.runWhenStart = v
	}
}

func WithRedirectStdErr(v string) Option {
	return func(c *config) {
		c.redirectStdErr = v
	}
}

func WithRedirectStdOut(v string) Option {
	return func(c *config) {
		c.redirectStdOut = v
	}
}
