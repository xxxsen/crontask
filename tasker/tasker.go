package tasker

import (
	"bytes"
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xxxsen/common/cmder"
	"github.com/xxxsen/common/logutil"
	"github.com/xxxsen/common/trace"
	"go.uber.org/zap"
)

type Tasker struct {
	c    *config
	name string
	id   uint64
	//
	lck       sync.Mutex
	isRunning atomic.Value
}

func NewTasker(name string, opts ...Option) (*Tasker, error) {
	if len(name) == 0 {
		name = "default"
	}
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.prgs) == 0 {
		return nil, fmt.Errorf("nil program")
	}
	if len(c.expression) == 0 {
		return nil, fmt.Errorf("nil cron expression")
	}
	return &Tasker{name: name, c: c}, nil
}

func (t *Tasker) Run(ctx context.Context) error {
	t.isRunning.Store(false)
	task := t.wrapTask(ctx)
	if t.c.runWhenStart {
		task()
	}
	cr := cron.New()
	_, err := cr.AddFunc(t.c.expression, task)
	if err != nil {
		return fmt.Errorf("add cron task fail, err:%w", err)
	}
	cr.Run()
	return nil
}

func (t *Tasker) wrapTask(c context.Context) func() {
	return func() {
		id := atomic.AddUint64(&t.id, 1)
		ctx := trace.WithTraceId(c, fmt.Sprintf("%s-%d", t.name, id))
		start := time.Now()
		if err := t.runTask(ctx, id); err != nil {
			logutil.GetLogger(ctx).Info("run job failed", zap.Error(err), zap.Duration("cost", time.Since(start)))
			return
		}
		logutil.GetLogger(ctx).Info("run job succ", zap.Duration("cost", time.Since(start)))
	}
}

func (t *Tasker) runTask(ctx context.Context, id uint64) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logutil.GetLogger(ctx).Error("run task cause panic", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
			err = fmt.Errorf("run task panic, panic:%v, stack:%s", r, string(debug.Stack()))
			return
		}
	}()
	if !t.isRunning.Load().(bool) {
		t.lck.Lock()
		if t.isRunning.Load().(bool) {
			logutil.GetLogger(ctx).
				Error("previous task still running, skip current task", zap.Uint64("current_id", id))
			t.lck.Unlock()
			return nil
		}
		t.isRunning.Store(true)
		t.lck.Unlock()
	}
	//
	err = t.runPrograms(ctx, id, t.c.prgs)
	//
	t.lck.Lock()
	t.isRunning.Store(false)
	t.lck.Unlock()
	return err
}

func (t *Tasker) runPrograms(ctx context.Context, id uint64, ps []prg) error {
	logger := logutil.GetLogger(ctx).With(zap.String("task", t.name), zap.Uint64("id", id))
	start := time.Now()
	for idx, p := range ps {
		if err := t.runProgram(ctx, id, &p); err != nil {
			logger.Error("exec sub task failed, skip next", zap.String("remark", p.remark), zap.Error(err))
			return fmt.Errorf("step:%d exec failed, err:[%w]", idx, err)
		}
	}
	logger.Info("task exec succ", zap.Duration("cost", time.Since(start)))
	return nil
}

func (t *Tasker) runProgram(ctx context.Context, id uint64, p *prg) error {
	logger := logutil.GetLogger(ctx).With(zap.Uint64("id", id), zap.String("remark", p.remark))
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("run program cause panic", zap.Any("err", rec), zap.String("stack", string(debug.Stack())))
			return
		}
	}()
	runner := cmder.New(p.workdir)
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	runner.SetOutput(&stdout, &stderr)
	if err := runner.Run(ctx, p.cmd, p.args...); err != nil {
		return fmt.Errorf("run cmd failed, cmd:%s, err:%w, errmsg:%s", p.cmd, err, stderr.String())
	}
	logutil.GetLogger(ctx).Info("run cmd succ", zap.String("cmd", p.cmd), zap.String("stdout", stdout.String()), zap.String("stderr", stderr.String()))
	return nil
}
