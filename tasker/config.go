package tasker

type prg struct {
	remark  string
	cmd     string
	args    []string
	workdir string
}

type config struct {
	prgs           []prg
	expression     string
	runWhenStart   bool
	redirectStdOut string
	redirectStdErr string
	tz             string
}

type Option func(c *config)

func WithAddProgram(remark string, workdir string, cmd string, args []string) Option {
	return func(c *config) {
		c.prgs = append(c.prgs, prg{
			remark:  remark,
			cmd:     cmd,
			args:    args,
			workdir: workdir,
		})
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

func WithTZ(v string) Option {
	return func(c *config) {
		c.tz = v
	}
}
