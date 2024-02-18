package tasker

import (
	"context"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xxxsen/common/cmder"
	"github.com/xxxsen/common/errs"
	"github.com/xxxsen/common/logutil"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultRedirectLogSize      = 10 * 1024 * 1024
	defaultRedirectLogFileCount = 5
	defaultRedirectLogKeepDays  = 7
)

type Tasker struct {
	c *config

	id uint64
	//
	lck       sync.Mutex
	isRunning atomic.Value
}

func NewTasker(opts ...Option) (*Tasker, error) {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.prgs) == 0 {
		return nil, errs.New(errs.ErrParam, "nil program")
	}
	if len(c.expression) == 0 {
		return nil, errs.New(errs.ErrParam, "nil cron expression")
	}
	return &Tasker{c: c}, nil
}

func (t *Tasker) Run() error {
	if t.c.runWhenStart {
		t.runPrograms(0, t.c.prgs)
	}

	t.isRunning.Store(false)
	crOpts := []cron.Option{}
	if len(t.c.tz) > 0 {
		loc, err := time.LoadLocation(t.c.tz)
		if err != nil {
			return errs.Wrap(errs.ErrParam, "parse time location fail", err)
		}
		crOpts = append(crOpts, cron.WithLocation(loc))
	}
	cr := cron.New(crOpts...)
	_, err := cr.AddFunc(t.c.expression, t.task)
	if err != nil {
		return errs.Wrap(errs.ErrServiceInternal, "add cron task fail", err)
	}
	cr.Run()
	return nil
}

func (t *Tasker) task() {
	id := atomic.AddUint64(&t.id, 1)
	if !t.isRunning.Load().(bool) {
		t.lck.Lock()
		if t.isRunning.Load().(bool) {
			logutil.GetLogger(context.Background()).
				Error("previous task still running, skip current task", zap.Uint64("current_id", id))
			t.lck.Unlock()
			return
		}
		t.isRunning.Store(true)
		t.lck.Unlock()
	}
	//
	t.runPrograms(id, t.c.prgs)
	//
	t.lck.Lock()
	t.isRunning.Store(false)
	t.lck.Unlock()
}

func (t *Tasker) defaultStreamByPath(loc string) io.Writer {
	logger := &lumberjack.Logger{
		// 日志输出文件路径
		Filename:   loc,
		MaxSize:    defaultRedirectLogSize / 1024 / 1024, // megabytes
		MaxBackups: defaultRedirectLogFileCount,
		MaxAge:     defaultRedirectLogKeepDays, //days
		Compress:   false,                      // disabled by default
	}
	return logger
}

func (t *Tasker) createStdOutStream() io.Writer {
	if len(t.c.redirectStdOut) > 0 {
		return t.defaultStreamByPath(t.c.redirectStdOut)
	}
	return os.Stdout
}

func (t *Tasker) createStdErrStream() io.Writer {
	if len(t.c.redirectStdErr) > 0 {
		return t.defaultStreamByPath(t.c.redirectStdErr)
	}
	return os.Stderr
}

func (t *Tasker) runPrograms(id uint64, ps []prg) {
	logger := logutil.GetLogger(context.Background()).With(zap.Uint64("id", id))
	for _, p := range t.c.prgs {
		if err := t.runProgram(id, &p); err != nil {
			logger.Error("exec sub task failed, skip next", zap.String("remark", p.remark), zap.Error(err))
			return
		}
	}
	logger.Info("task exec succ")
}

func (t *Tasker) runProgram(id uint64, p *prg) error {
	logger := logutil.GetLogger(context.Background()).With(zap.Uint64("id", id), zap.String("remark", p.remark))
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("run sub task cause panic", zap.Any("err", rec), zap.String("stack", string(debug.Stack())))
			return
		}
	}()
	runner := cmder.NewCMD(p.workdir)
	runner.SetOutput(t.createStdOutStream(), t.createStdErrStream())
	now := time.Now()
	if err := runner.Run(context.Background(), p.cmd, p.args...); err != nil {
		return err
	}
	logger.Debug("run sub task succ", zap.Duration("cost", time.Since(now)))
	return nil
}
