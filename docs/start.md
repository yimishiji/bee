

开始
===

### 创建新的项目，app-test这项目名称
bee api app-test

```$xslt
       create src/app-test
       create src/app-test/main.go
       create src/app-test/conf
       create src/app-test/controllers
       create src/app-test/tests
       create src/app-test/conf/app.conf
       create src/app-test/models
       create src/app-test/routers/
       create src/app-test/routers/router.go
       create src/app-test/controllers/object.go
       create src/app-test/controllers/user.go
       create src/app-test/tests/default_test.go
       create src/app-test/models/object/model.go
       create src/app-test/models/user/model.go
       create src/app-test/pkg/middle-wares/middleware.go
       create src/app-test/pkg/middle-wares/allow-all.go
       create src/app-test/pkg/middle-wares/allow-token.go
       create src/app-test/service-logics/user/user.go
       create src/app-test/service-logics/health-checks/database.go
       create src/app-test/service-logics/health-checks/redis.go
       create src/app-test/bee.json
       create src/app-test/.gitignore
       create src/app-test/Dockerfile
       create src/app-test/.dockerignore

```

### 项目基本目录结构说明

入口文件
>- /main.go 

配置目录
>- conf
>- /conf/app.conf
>- /conf/app_local.conf         本地化配置

测试代码
>- /tests
>- /tests/default_test.go

路由层
>- /routers
>- /routers/router.go

控制器层
>- /controllers
>- /controllers/object.go
>- /controllers/user.go

模型层
>- /models
>- /models/object/model.go
>- /models/user/model.go

 模型对应的表结构
>- /models/table-structs

pkg其它类包
>- /pkg/middle-wares/middleware.go           中间件
>- /pkg/middle-wares/allow-all.go            开放访问的请求配置
>- /pkg/middle-wares/allow-token.go          登录用户即可访问的请求配置

业务逻辑层
>- /service-logics                                    
>- /service-logics/user/user.go               
   
健康检察验证器
>- /service-logics/health-checks                  
>- /service-logics/health-checks/database.go  数据库连接检查
>- /service-logics/health-checks/redis.go     redis连接检查        

beego框架bee工具配置文件
>- bee.json

git版本控制忽略文件列表
>- .gitignore

docker描述文件,及docker编译时的忽略文件列表
>- Dockerfile
>- .dockerignore

### generate appcode 自动生成代码工具改造，
[generate appcode -level=4](generate-appcode.md)

### 定时任务与队列
[crontab-and-queue.md](crontab-and-queue.md)