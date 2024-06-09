crontask
===


定时任务工具, 用于方便地在docker中执行定时任务。

## 使用env运行

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

## 使用配置文件运行

基础配置模板:

```json
{
    "log": {
        "level": "debug",
        "console": true
    },
    "tz": "Asia/Shanghai", //timezone
    "run_when_start": true, //是否启动的时候执行一次
    "task_name": "default", //任务名
    "crontask_expression": "*/1 * * * *",
    "programs": [ //支持多个任务, 按列表顺序进行执行。
        {
            "remark": "t1",
            "cmd": "/usr/bin/ls",
            "args": [
                "-alh"
            ]
        },
        {
            "remark": "t2",
            "cmd": "/usr/bin/echo",
            "args": [
                "hahaha"
            ]
        }
    ],
    "notify": { //结束完成后进行通知
        "succ": {
            "cmd": "/usr/bin/echo", //执行成功时运行的命令
            "args": [
            ]
        },
        "fail": {
            "cmd": "/usr/bin/echo", //执行失败时运行的命令
            "args": [
            ]
        },
        "finish": {
            "cmd": "/usr/bin/echo", //执行结束时运行的命令
            "args": [
            ]
        }
    },
    "redirect_stdout": "/tmp/test_out", //将输出结果进行重定向
    "redirect_stderr": "/tmp/test_err"
}
```

通知占位参数: 

主要用于对通知的参数进行重写, 目前支持以下占位符。

|占位符|说明|
|---|---|
|TASK_RUN_ID|当前任务执行的批次|
|TASK_NAME|任务名|
|TASK_SUCC|当前任务是否执行成功|
|TASK_ERRMSG|当前任务的错误信息, 仅失败的时候有值|
|TASK_RUN_TIME|任务执行耗时|

占位符使用方式:

```json
{
    "programs": [...],
    "notify": {
        "succ": {
            "cmd": "/usr/bin/echo",
            "args": [
                "task exec succ, runid:{{TASK_RUN_ID}}, task_name:{{TASK_NAME}}, cost:{{TASK_RUN_TIME}}"
            ]
        }
    }
}

```