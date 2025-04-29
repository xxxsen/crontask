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

func runWithConfig(c *config.Config) {
	logger := logger.Init(c.Log.File, c.Log.Level, int(c.Log.FileCount), int(c.Log.FileSize), int(c.Log.KeepDays), c.Log.Console)
	logger.Info("recv config", zap.Any("config", *c))

	opts := []tasker.Option{
		tasker.WithCronExpression(c.CrontaskExpression),
		tasker.WithRunWhenStart(c.RunWhenStart),
	}
	for _, p := range c.Programs {
		opts = append(opts, tasker.WithAddProgram(p.Remark, p.WorkDir, p.Cmd, p.Args))
	}
	tk, err := tasker.NewTasker(c.TaskName, opts...)
	if err != nil {
		logger.Fatal("create tasker fail", zap.Error(err))
	}
	logger.Info("create tasker success, start it...")
	if err := tk.Run(context.Background()); err != nil {
		logger.Fatal("run tasker fail", zap.Error(err))
	}
}
