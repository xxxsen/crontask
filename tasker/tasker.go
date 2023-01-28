package tasker

import (
	"context"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/robfig/cron/v3"
	"github.com/xxxsen/common/cmder"
	"github.com/xxxsen/common/errs"
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
	if len(c.executor) == 0 {
		return nil, errs.New(errs.ErrParam, "nil program name")
	}
	if len(c.expression) == 0 {
		return nil, errs.New(errs.ErrParam, "nil cron expression")
	}
	return &Tasker{c: c}, nil
}

func (t *Tasker) Run() error {
	if t.c.runWhenStart {
		t.runWithLockCheck(0)
	}

	t.isRunning.Store(false)
	cr := cron.New()
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
			log.Printf("previous task still running, skip current task, current id:%d", id)
			t.lck.Unlock()
			return
		}
		t.isRunning.Store(true)
		t.lck.Unlock()
	}
	//
	t.runWithLockCheck(id)
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
	if len(t.c.redirectStdOut) == 0 {
		return os.Stdout
	}
	return t.defaultStreamByPath(t.c.redirectStdOut)
}

func (t *Tasker) createStdErrStream() io.Writer {
	if len(t.c.redirectStdErr) == 0 {
		return os.Stderr
	}
	return t.defaultStreamByPath(t.c.redirectStdErr)
}

func (t *Tasker) runWithLockCheck(id uint64) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("run task cause panic, id:%d, err:%v, stack:%s", id, err, string(debug.Stack()))
			return
		}
	}()
	runner := cmder.NewCMD(t.c.workdir)
	if t.c.enableUserCred {
		runner.SetID(uint32(t.c.uid), uint32(t.c.gid))
	}
	runner.SetOutput(t.createStdOutStream(), t.createStdErrStream())
	if err := runner.Run(context.Background(), t.c.executor, t.c.params...); err != nil {
		log.Printf("id:%d task exec fail, err:%v", id, err)
		return
	}
	log.Printf("id:%d task exec succ", id)

}
