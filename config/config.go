package config

import (
	"encoding/json"
	"os"

	"github.com/xxxsen/common/logger"
)

type Program struct {
	Remark  string   `json:"remark"`
	WorkDir string   `json:"workdir"`
	Cmd     string   `json:"cmd"`
	Args    []string `json:"args"`
}

type Config struct {
	Log logger.LogConfig `json:"log"`
	*TaskConfig
	TaskList []TaskConfig `json:"task_list"`
}

type TaskConfig struct {
	TaskName string `json:"task_name"`
	Expr     string `json:"expr"`
	//Deprecated: use Expr instead
	CrontaskExpression string    `json:"crontask_expression"`
	Programs           []Program `json:"programs"`
	RunWhenStart       bool      `json:"run_when_start"`
}

func Parse(file string) (*Config, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err = json.Unmarshal(raw, c); err != nil {
		return nil, err
	}
	if c.TaskConfig != nil {
		c.TaskList = append(c.TaskList, *c.TaskConfig)
	}
	for _, item := range c.TaskList {
		if item.Expr == "" {
			item.Expr = item.CrontaskExpression
		}
	}
	return c, nil

}
