crontask
===


定时任务工具, 用于方便地在docker中执行定时任务。

env配置:

| key                    | type   | 备注                                     |
| ---------------------- | ------ | ---------------------------------------- |
| CRONTASK_EXPRESSION    | string | cron表达式                               |
| ENABLE_USER_GROUP_SPEC | bool   | 启用后会以指定的uid、gid执行定时任务程序 |
| UID                    | int    | uid                                      |
| GID                    | int    | gid                                      |

运行:

```shell=
CRONTASK_EXPRESSION="${CRON_EXPRESSION}" crontask ${command} [${arg1}, ${arg2}, ...]

# 例子
CRONTASK_EXPRESSION='*/1 * * * *' crontask ls -alh --color
```