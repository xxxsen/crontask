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
    "task_list": [
        {
            "task_name": "in_task_list_1", //任务名
            "expr": "*/1 * * * *", //cron表达式
            "run_when_start": true, //启动的时候预先执行一次
            "programs": [ //子任务列表(可以配置多个)
                {
                    "remark": "echo 1", 
                    "workdir": "", //工作目录
                    "cmd": "echo",
                    "args": [
                        "1"
                    ]
                }
            ]
        },
        {
            "task_name": "in_task_list_2",
            "expr": "*/1 * * * *",
            "run_when_start": true,
            "programs": [
                {
                    "remark": "echo 2",
                    "workdir": "",
                    "cmd": "echo",
                    "args": [
                        "2"
                    ]
                }
            ]
        }
    ]
}
```
