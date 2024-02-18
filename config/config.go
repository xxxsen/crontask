package config

import (
	"encoding/json"
	"os"

	"github.com/xxxsen/common/logger"
)

type Program struct {
	Remark  string   `json:"remark"`
	WorkDir string   `json:"work_dir"`
	Cmd     string   `json:"cmd"`
	Args    []string `json:"args"`
}

type Config struct {
	Log                logger.LogConfig `json:"log"`
	TZ                 string           `json:"tz"`
	CrontaskExpression string           `json:"crontask_expression"`
	Programs           []Program        `json:"programs"`
	RunWhenStart       bool             `json:"run_when_start"`
	RedirectStdout     string           `json:"redirect_stdout"`
	RedirectStderr     string           `json:"redirect_stderr"`
}

func Parse(file string) (*Config, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c := &Config{
		TZ: "Asia/Shanghai",
	}
	if err = json.Unmarshal(raw, c); err != nil {
		return nil, err
	}
	return c, nil

}
