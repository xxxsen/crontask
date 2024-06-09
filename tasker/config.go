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
	onSucc         *prg
	onFail         *prg
	onFinish       *prg
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

func WithSuccNotify(cmd string, args []string) Option {
	return func(c *config) {
		c.onSucc = &prg{
			remark: "on_succ",
			cmd:    cmd,
			args:   args,
		}
	}
}

func WithFailNotify(cmd string, args []string) Option {
	return func(c *config) {
		c.onFail = &prg{
			remark: "on_fail",
			cmd:    cmd,
			args:   args,
		}
	}
}

func WithFinishNotify(cmd string, args []string) Option {
	return func(c *config) {
		c.onFinish = &prg{
			remark: "on_finish",
			cmd:    cmd,
			args:   args,
		}
	}
}
