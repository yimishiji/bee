

在原生的bee工具做了改造，适应手复杂的业务场景
===

## 配置 bee 脚本配置
src/appname/bee.json
```$xslt
    {
        "version": 0,
        "database": {
            "driver": "mysql",
            "conn": "username:password(localhost:3306)/dbname",
        },
        "cmd_args": [],
        "enable_reload": true
    }
```


##  生成app代码
项目目录下执行以下代码将生成api-vue代码
```$xslt
/gopath/src/monitor-api>bee generate appcode -tables="member_coupon,member_deposit_logs" -level=4
______
| ___ \
| |_/ /  ___   ___
| ___ \ / _ \ / _ \
| |_/ /|  __/|  __/
\____/  \___| \___| v1.10.0
14:46:24 INFO     ▶ 0001 Using 'mysql' as 'SQLDriver'
14:46:24 INFO     ▶ 0002 Using 'bpm-api:Ay4tghdU@tcp(yimishiji-mysql.com:3306)/ms_bpm' as 'SQLConn'
14:46:24 INFO     ▶ 0003 Using 'member_coupon,member_deposit_logs' as 'Tables'
14:46:24 INFO     ▶ 0004 Using '4' as 'Level'
14:46:24 INFO     ▶ 0005 Analyzing database tables...
14:46:24 INFO     ▶ 0006 Creating model files...
        create   models\member\coupon\model.go
        create   models\table-structs\member-coupon.go
        create   models\member\deposit-logs\model.go
        create   models\table-structs\member-deposit-logs.go
14:46:25 INFO     ▶ 0007 Creating controller files...
        create   controllers\member-coupon.go
        create   controllers\member-deposit-logs.go
14:46:25 INFO     ▶ 0008 Creating filter files...
        create   filters\member\coupon\input.go
        create   filters\member\deposit-logs\input.go
14:46:25 INFO     ▶ 0009 Creating router files...
14:46:25 WARN     ▶ 0010 Skipped create file 'routers\router.go'
14:46:25 INFO     ▶ 0011 Creating vue files...
        create   vue\src\components\member-coupon\index.vue
        create   vue\src\components\member-coupon\create-component.vue
        create   vue\src\components\member-coupon\edit-component.vue
        create   vue\src\components\member-coupon\colsetting-component.vue
        create   vue\src\components\member-deposit-logs\index.vue
        create   vue\src\components\member-deposit-logs\create-component.vue
        create   vue\src\components\member-deposit-logs\edit-component.vue
        create   vue\src\components\member-deposit-logs\colsetting-component.vue
14:46:25 WARN     ▶ 0012 add to file this route
 add to operate list:

        operateList = append(operateList, RoleRight{
                RightAction: "[GET]/member-coupon",
        })
        operateList = append(operateList, RoleRight{
                RightAction: "[POST]/member-coupon",
        })
        operateList = append(operateList, RoleRight{
                RightAction: "[PUT]/member-coupon",
        })
        operateList = append(operateList, RoleRight{
                RightAction: "[DELETE]/member-coupon",
        })

        operateList = append(operateList, RoleRight{
                RightAction: "[GET]/member-deposit-logs",
        })
        operateList = append(operateList, RoleRight{
                RightAction: "[POST]/member-deposit-logs",
        })
        operateList = append(operateList, RoleRight{
                RightAction: "[PUT]/member-deposit-logs",
        })
        operateList = append(operateList, RoleRight{
                RightAction: "[DELETE]/member-deposit-logs",
        })

add to routers/router.go

        beego.NSNamespace("/member-coupon",
                beego.NSInclude(
                        &controllers.MemberCouponController{},
                ),
        ),
        beego.NSNamespace("/member-deposit-logs",
                beego.NSInclude(
                        &controllers.MemberDepositLogsController{},
                ),
        ),
add to vue vue/src/router/index.js

        {
            path: '/member-coupon/index',
            component: name => require(['../components/member-coupon/Index'], name),
        },
        {
            path: '/member-deposit-logs/index',
            component: name => require(['../components/member-deposit-logs/Index'], name),
        },
add to vue menu

        {"name":"MemberCoupon","url":"/member-coupon/index","icon":"bars"},
        {"name":"MemberDepositLogs","url":"/member-deposit-logs/index","icon":"bars"},

14:46:25 SUCCESS  ▶ 0013 Appcode successfully generated!
```

### model模型层
- 表结构层，申明表字段，表名。
    models\table-structs\member-coupon.go
    models\table-structs\member-deposit-logs.go
- 一般model，与表结构层一对一继承，可扩展些层，可多个一般model层对应同一个表结构层。可申明关连关系，关连表时可对应关连多个表结构层，也可以关连其它一般model层。一般model会自动分组，表以下划线分隔，如果第一部分相同，则会分到同一目录下
    * models\member\coupon\model.go
    * models\member\deposit-logs\model.go
- 一般model，与表结构层一对一继承，可扩展些层，可多个一般model层对应同一个表结构层。
- 可声明关连关系，关连表时可对应关连多个表结构层，也可以关连其它一般model层。
- 一般model会自动分组，表以下划线分隔，如果第一部分相同，则会分到同一目录下。
- model层 关连原理 [generate-appcode.md](generate-appcode.md)

### controller控制器层
- 申请filter过虑器
```$xslt
    import (
        MemberDepositLogsFilter "api-test/filters/member/deposit-logs"
        MemberDepositLogsModel "api-test/models/member/deposit-logs"
    )
    
    // MemberDepositLogsController operations for MemberDepositLogs
    type MemberDepositLogsController struct {
        base.Controller
        filter *MemberDepositLogsFilter.Filter
    }
```
- 初始化filter过滤器
```$xslt
    // init inputFilter
    func (c *MemberDepositLogsController) Prepare() {
        c.filter = MemberDepositLogsFilter.NewFilter(c.Ctx.Input)
    }
```
- 

### filter过滤层
- 接收接口请求的数据，解析成可以直接调用的struct,传给controller层。
- 不同的表单数据声明不同的结构体，做不同在验证
- 一般model会自动分组，表以下划线分隔，如果第一部分相同，则会分到同一目录下。


### vue页面
- index.vue  列表页
- create-component.vue  创建组件
- edit-component.vue  编辑/查看组件
- colsetting-component.vue //列表页列显示设置组件

### 自动生成bee操作权限项
- 将生成的操作权限列表加入到service-logics/user/user.go 的 GetOperateListByAccesstoken方法中，
```$xslt
    operateList = append(operateList, RoleRight{
        RightAction: "[GET]/member-coupon",
    })
```
- 方便开发环境调试，登录用户就享有这些新生成的权限项，
- 后续需将些权限项添加到权限系统，非开发环境分配后才有权限访问

### 自动生成bee框架的路由规则
- 将生成的路由规则加入到/routers/router.go 路由文件init方法中
```$xslt
    beego.NSNamespace("/member-coupon",
        beego.NSInclude(
             &controllers.MemberCouponController{},
        ),
    ),
```

### 自动生成vue框架的路由规则
- 将生成的路由规则加入到 /vue/src/router/index.js 路由文件中
```$xslt
    {
        path: '/member-coupon/index',
        component: name => require(['../components/member-coupon/Index'], name),
    },
```

### 自动生成vue框架菜单规则
- 自动生成的菜单项 数据类型兼容v-menu菜单组件
```$xslt
    {"name":"MemberCoupon","url":"/member-coupon/index","icon":"bars"},
```