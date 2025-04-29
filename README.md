crontask
===

定时任务工具, 用于方便地在docker中执行定时任务。

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
    ]
}
```
