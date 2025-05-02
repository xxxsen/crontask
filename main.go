package main

import (
	"context"
	"crontask/config"
	"crontask/tasker"
	"flag"
	"log"

	"github.com/xxxsen/common/logger"
	"go.uber.org/zap"
)

var conf = flag.String("config", "./config.json", "config file")

func main() {
	flag.Parse()
	c, err := config.Parse(*conf)
	if err != nil {
		log.Fatalf("parse config failed, err:%v", err)
	}
	runWithConfig(c)
}

func generateJobConfigs(c *config.Config) []*tasker.JobConfig {
	rs := make([]*tasker.JobConfig, 0, len(c.TaskList))
	for _, tk := range c.TaskList {
		jb := &tasker.JobConfig{
			Name:         tk.TaskName,
			Expr:         tk.Expr,
			RunWhenStart: tk.RunWhenStart,
		}
		for _, p := range tk.Programs {
			jb.SubTasks = append(jb.SubTasks, &tasker.TaskConfig{
				Remark:  p.Remark,
				WorkDir: p.WorkDir,
				Cmd:     p.Cmd,
				Args:    p.Args,
			})
		}
		rs = append(rs, jb)
	}
	return rs
}

func runWithConfig(c *config.Config) {
	logger := logger.Init(c.Log.File, c.Log.Level, int(c.Log.FileCount), int(c.Log.FileSize), int(c.Log.KeepDays), c.Log.Console)
	logger.Info("recv config", zap.Any("config", *c))

	tk := tasker.New()
	jbs := generateJobConfigs(c)
	for _, jc := range jbs {
		tk.AddJob(jc)
	}
	logger.Info("create tasker success, start it...", zap.Int("job_count", len(jbs)))
	if err := tk.Run(context.Background()); err != nil {
		logger.Fatal("run tasker fail", zap.Error(err))
	}
}
