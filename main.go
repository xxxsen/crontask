package main

import (
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
	log.Printf("config init succ, c:%+v", *c)

	logger := logger.Init(c.Log.File, c.Log.Level, int(c.Log.FileCount), int(c.Log.FileSize), int(c.Log.KeepDays), c.Log.Console)

	opts := []tasker.Option{
		tasker.WithCronExpression(c.CronExpression),
		tasker.WithTZ(c.TZ),
		tasker.WithRunWhenStart(c.RunWhenStart),
	}
	for _, p := range c.Programs {
		opts = append(opts, tasker.WithAddProgram(p.Remark, p.WorkDir, p.Cmd, p.Args))
	}
	if len(c.RedirectStderr) > 0 {
		opts = append(opts, tasker.WithRedirectStdErr(c.RedirectStderr))
	}
	if len(c.RedirectStdout) > 0 {
		opts = append(opts, tasker.WithRedirectStdOut(c.RedirectStdout))
	}
	tk, err := tasker.NewTasker(opts...)
	if err != nil {
		logger.Fatal("create tasker fail", zap.Error(err))
	}
	if err := tk.Run(); err != nil {
		logger.Fatal("run tasker fail", zap.Error(err))
	}
}
