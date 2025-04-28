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
	"github.com/xxxsen/common/replacer"
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

func (t *Tasker) Run() error {
	t.isRunning.Store(false)
	if t.c.runWhenStart {
		t.task()
	}
	cr := cron.New()
	_, err := cr.AddFunc(t.c.expression, t.task)
	if err != nil {
		return fmt.Errorf("add cron task fail, err:%w", err)
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
	start := time.Now()
	err := t.runPrograms(id, t.c.prgs)
	t.runNotify(id, t.name, time.Since(start), err)
	//
	t.lck.Lock()
	t.isRunning.Store(false)
	t.lck.Unlock()
}

func (t *Tasker) runNotify(id uint64, name string, cost time.Duration, err error) {
	logger := logutil.GetLogger(context.Background()).With(zap.String("task", name), zap.Uint64("run_id", id))
	runNotify := make([]*prg, 0, 3)
	if t.c.onFinish != nil {
		runNotify = append(runNotify, t.c.onFinish)
	}
	if t.c.onFail != nil && err != nil {
		runNotify = append(runNotify, t.c.onFail)
	}
	if t.c.onSucc != nil && err == nil {
		runNotify = append(runNotify, t.c.onSucc)
	}
	for _, nt := range runNotify {
		args := t.rewriteNotifyArgs(nt.args, id, name, cost, err)
		if err := t.runProgram(id, &prg{
			remark:  nt.remark,
			cmd:     nt.cmd,
			args:    args,
			workdir: nt.workdir,
		}); err != nil {
			logger.Error("run notify failed", zap.Error(err), zap.String("remark", nt.remark))
		}
	}
}

func (t *Tasker) rewriteNotifyArgs(inputArgs []string, id uint64, name string, cost time.Duration, err error) []string {
	outputArgs := make([]string, 0, len(inputArgs))
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	param := map[string]interface{}{
		KeyRunID:       id,
		KeyTaskName:    name,
		KeyTaskSucc:    err == nil,
		KeyTaskRunTime: cost.Milliseconds(),
		KeyTaskErrMsg:  errMsg,
	}
	for _, arg := range inputArgs {
		arg = replacer.ReplaceByMap(arg, param)
		outputArgs = append(outputArgs, arg)
	}
	return outputArgs
}

func (t *Tasker) runPrograms(id uint64, ps []prg) error {
	logger := logutil.GetLogger(context.Background()).With(zap.String("task", t.name), zap.Uint64("id", id))
	start := time.Now()
	for idx, p := range ps {
		if err := t.runProgram(id, &p); err != nil {
			logger.Error("exec sub task failed, skip next", zap.String("remark", p.remark), zap.Error(err))
			return fmt.Errorf("step:%d exec failed, err:[%w]", idx, err)
		}
	}
	logger.Info("task exec succ", zap.Duration("cost", time.Since(start)))
	return nil
}

func (t *Tasker) runProgram(id uint64, p *prg) error {
	logger := logutil.GetLogger(context.Background()).With(zap.Uint64("id", id), zap.String("remark", p.remark))
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
	if err := runner.Run(context.Background(), p.cmd, p.args...); err != nil {
		return fmt.Errorf("run cmd failed, cmd:%s, err:%w, errmsg:%s", p.cmd, err, stderr.String())
	}
	logutil.GetLogger(context.Background()).Info("run cmd succ", zap.String("cmd", p.cmd), zap.String("stdout", stdout.String()), zap.String("stderr", stderr.String()))
	return nil
}
