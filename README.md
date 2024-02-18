crontask
===


定时任务工具, 用于方便地在docker中执行定时任务。

env配置:

| key                    | type   | 备注                                     |
| ---------------------- | ------ | ---------------------------------------- |
| CRONTASK_EXPRESSION    | string | cron表达式                               |
| RUN_WHEN_START         | bool   | 开启后, 在程序启动的时候会先执行一次任务 |
| REDIRECT_CMD_STDOUT    | string | 将被执行命令的输出流重定向到指定文件     |
| REDIRECT_CMD_STDERR    | string | 将被执行命令的错误流重定向到指定文件     |
| TZ                     | string | 指定时区                                 |

运行:

```shell=
CRONTASK_EXPRESSION="${CRON_EXPRESSION}" crontask ${command} [${arg1}, ${arg2}, ...]

# 例子
CRONTASK_EXPRESSION='*/1 * * * *' crontask ls -alh --color
```