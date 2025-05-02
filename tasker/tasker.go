package tasker

import (
	"bytes"
	"context"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xxxsen/common/cmder"
	"github.com/xxxsen/common/logutil"
	"github.com/xxxsen/common/trace"
	"go.uber.org/zap"
)

type tasker struct {
	jobs []*JobConfig
}

type ITasker interface {
	Run(ctx context.Context) error
	AddJob(jb *JobConfig)
}

func New() ITasker {
	return &tasker{}
}

func (t *tasker) AddJob(jb *JobConfig) {
	t.jobs = append(t.jobs, jb)
}

func (t *tasker) Run(ctx context.Context) error {
	cr := cron.New()
	for _, jb := range t.jobs {
		if err := t.initAndAddJob(ctx, cr, jb); err != nil {
			return fmt.Errorf("init job %s failed, err:%w", jb.Name, err)
		}
	}
	cr.Run()
	return nil
}

func (t *tasker) initAndAddJob(ctx context.Context, cr *cron.Cron, jb *JobConfig) error {
	if jb.Expr == "" {
		return fmt.Errorf("job %s has no cron expression", jb.Name)
	}
	task := t.wrapTask(ctx, jb)
	if jb.RunWhenStart {
		task()
	}
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	s, err := p.Parse(jb.Expr)
	if err != nil {
		return fmt.Errorf("parse cron expr failed, err:%w", err)
	}
	next := s.Next(time.Now())
	logutil.GetLogger(ctx).Info("add job",
		zap.String("name", jb.Name),
		zap.String("expr", jb.Expr),
		zap.Bool("run_when_start", jb.RunWhenStart),
		zap.Int("sub_task_count", len(jb.SubTasks)),
		zap.Time("next_run_time", next),
	)
	if _, err := cr.AddFunc(jb.Expr, task); err != nil {
		return fmt.Errorf("add job %s failed, err:%w", jb.Name, err)
	}
	return nil
}

func (t *tasker) wrapTask(c context.Context, jb *JobConfig) func() {
	var id uint64
	return func() {
		id := atomic.AddUint64(&id, 1)
		ctx := trace.WithTraceId(c, fmt.Sprintf("%s-%d", jb.Name, id))
		start := time.Now()
		if err := t.runTask(ctx, jb); err != nil {
			logutil.GetLogger(ctx).Info("run job failed", zap.Error(err), zap.Duration("cost", time.Since(start)))
			return
		}
		logutil.GetLogger(ctx).Info("run job succ", zap.Duration("cost", time.Since(start)))
	}
}

func (t *tasker) runTask(ctx context.Context, jb *JobConfig) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("run task panic, panic:%v, stack:%s", r, string(debug.Stack()))
			return
		}
	}()
	err = t.runPrograms(ctx, jb.SubTasks)
	if err != nil {
		return err
	}
	return nil
}

func (t *tasker) runPrograms(ctx context.Context, tks []*TaskConfig) error {
	for idx, tk := range tks {
		start := time.Now()
		if err := t.runProgram(ctx, tk); err != nil {
			return fmt.Errorf("exec sub task failed, tidx:%d, remark:%s, err:%w", idx, tk.Remark, err)
		}
		logutil.GetLogger(ctx).Info("sub task exec succ", zap.Int("idx", idx), zap.String("remark", tk.Remark), zap.Duration("cost", time.Since(start)))
	}
	return nil
}

func (t *tasker) runProgram(ctx context.Context, tk *TaskConfig) error {
	runner := cmder.New(tk.WorkDir)
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	runner.SetOutput(&stdout, &stderr)
	if err := runner.Run(ctx, tk.Cmd, tk.Args...); err != nil {
		return fmt.Errorf("run cmd failed, cmd:%s, args:[%+v], err:%w, errmsg:%s", tk.Cmd, tk.Args, err, stderr.String())
	}
	logutil.GetLogger(ctx).Info("run cmd succ", zap.String("cmd", tk.Cmd), zap.Strings("args", tk.Args), zap.String("stdout", stdout.String()), zap.String("stderr", stderr.String()))
	return nil
}
