

定时任务与队列服务
===

### 程序以cron模式运行

- run-mode 参数，运行模式。如果不传默认为api, 仅为cron时才开启定时任务和队表服务进程
```
/gopath/src/api-test> go build

/gopath/src/api-test> ./api-test -run-mode=cron
 
```


### 框架入口main.go 接收run-mode
框架入口main.go 接收run-mode运行模式参数逻辑

```$xslt
func main() {
	runMode := flag.String("run-mode", "api", "app run mode")
	flag.Parse()
	//命令行模式
	if *runMode == "cron" {
		print("======crontab mode\n")
		commands.RunQueue()
		commands.RunCrontab()
	}
	...
}
```


### 定时任务 RunCrontab
定时任务初始化方法，所有的定时任务需要些配置
src/api-test/commands/crontab.go
```$xslt
package commands

import (
	InstanceCallback "bpm-api/commands/instance-callback"

	"github.com/astaxie/beego/toolbox"
)

//定时任务初始化
func RunCrontab() {
	//3分钟跑一次，一小时内的回调失败尝试
	toolbox.AddTask("tk1", toolbox.NewTask("tk1", "*/3 * * * * *", InstanceCallback.RunSecond))
	//10分钟跑一次，一天内的回调失败尝试
	toolbox.AddTask("tk1", toolbox.NewTask("tk1", "*/10 * * * * *", InstanceCallback.RunMinute))
	//1小时跑一次，3天内的回调失败尝试
	toolbox.AddTask("tk1", toolbox.NewTask("tk1", "1 */1 * * * *", InstanceCallback.RunHour))
}
```


### 队表消费进程初始化
申明消费进程
src/api-test/commands/queue.go
```$xslt
package commands

import InstanceCallbackConsumer "bpm-api/commands/consumers/instance-callback"

//队列初始化
func RunQueue() {
	//回调失败重试队列消费程序
	InstanceCallbackConsumer.Init()
}
```

### 消费进程实例
所有的消费程序放在 commands/consumers 目录下，每个消费进程单独设定一个包
例程 
src/api-test/commands/consumers/instance-callback/consumer.go
```$xslt
package InstanceCallbackConsumer

import (
	"bpm-api/models/WorkflowInstanceCallbackRecordModel"
	"time"

	"github.com/astaxie/beego"
)

var (
	//列表进程数
	Num = 5
	//队列通道，长度不限制
	Chan = make(chan int)
	//消息队列队协程数限制，最大同时运行的协程数
	Limit = make(chan bool, 1000)
)

func Init() {
	for i := 0; i < Num; i++ {
		go queue(i, Chan)
	}
}

//队列监听，消费
func queue(qid int, rchan chan int) {
	for {
		select {
		case id := <-rchan:
			//占用协程数，如果送到设定上限，后面的进程将等待处理
			Limit <- true
			
			beego.Info("InstanceCallbackConsumer queue run", qid, id)
			go tryCall(id)
		}
	}
}

//消费逻辑处理，注意释放协程占用
func tryCall(id int) {
	v, _ := WorkflowInstanceCallbackRecordModel.GetById(id)
	v.Call()
	v.UpdatedAt = int(time.Now().Unix())
	if err := WorkflowInstanceCallbackRecordModel.Update(&v); err != nil {
		beego.Error("update callback error,", err.Error())
	}

	//延时释放可用协程数
	defer func() {
		<-Limit
	}()
}

```
- Num 消费队列的进程数
- Chan 队列管道
- Limit 消息队列队协程数限制，最大同时运行的协程数


### 定时任务实例 
定时任务将要执行的数据加入队列通道
```$xslt
package InstanceCallback

import (
	"bpm-api/models/WorkflowInstanceCallbackRecordModel"

	InstanceCallbackConsumer "bpm-api/commands/consumers/instance-callback"

	"time"

	"github.com/astaxie/beego"
	"github.com/yimishiji/bee/pkg/db"
)

//一小时内，每3分钟尝试一次失败的
func RunSecond() error {
	beego.Info("RunSecond")
	var list []WorkflowInstanceCallbackRecordModel.Model
	err := db.Conn.Select("id").
		Where("created_at > ?", time.Now().Unix()-3600).
		Where("is_success = 0").Limit(10000).Find(&list).Error
	if err != nil {
		beego.Notice("callback runSecond skip", err.Error())
		return nil
	}

	//通道回调，队列异步处理
	for _, item := range list {
		InstanceCallbackConsumer.Chan <- item.Id
	}
	return nil
}
```
- 加入队列的用法

``` InstanceCallbackConsumer.Chan <- item.Id ```