
相对于原生beego, 我们做了什么？
===

## 调整框架结构

```
    conf                
    controllers           
    database              
    filters               接收参数过滤层
    models
    pkg                   基础包类，sdk
    routers     
    serviceLogics         逻辑处理层，逻辑层 
    tests
    venodr
    vue
```


## model 层改造

```$xslt
    modelsTableStructs  表结构层
        table1  
        table2
        table3
    Table1Mode1    表模型层
        model extends table1
    Table1Mode2    表模型层
        model extends table2
    Table1Mode3    表模型层
        model extends table3
    
```

#### 简单关连模型 单次关连，只关连其它表，关连的表不再关连其它的

```$xslt
    modelsTableStructs  表结构层
        table1  
        table2
        table3
    Table1Mode    表模型层
        model extends table1
              rels table2
              rels table3
```

#### 多重关连模型，需要建一个model的副本，实现正反向关连

```$xslt
    modelsTableStructs  表结构层
        table1  
        table2
        table3
      
    Table2Mode    表模型层
        model extends table2
              rels Table3Mode
              
    Table3Mode    表模型层
        model extends 
              rels Table2ModeSimple
              
    Table2ModeSimple    表模型层
        model extends table2
```


## 权限验证
```$xslt
   /pkg/middleWares/middleware.go 
   
   // 只验证登录，不验证授权的操作列表，登录即有权限的接口
   /pkg/middleWares/allAllowTokenUrlList.go
   
   // 不需要验证权限的操作列表
   /pkg/middleWares/noTokenUrlList.go
```


## 统一列表页 请求参数

src/github.com/yimishiji/bee/pkg/filters/PageCommonParams.go
```
    //公类页公共参数结构
    type PageCommonParams struct {
        Field      []string
        Sort       []string
        Orders     []string
        Querys     map[string]string
        Limits     int64
        Offsets    int64
        SortFields []string
        Rels       []string
    }
```
获取方法放在filter层，filter层还可以特定的查询条件做特殊的验证规则

```$xslt
    //分页参数
    func (this *WorkflowInputFieldsFilter) GetListPrams() (params *filters.PageCommonParams, err error) {
        if params, err := this.GetPagePublicParams(); err == nil {
    
            //验证筛选的条件合法性
            //if t, ok := params.Querys["type"]; ok {
            //	if filters.InStingArr(t, []string{"orders", "goods", "users"}) == false {
            //		return params, errors.New("type is not enable")
            //	}
            //}
    
            return params, nil
        } else {
            return params, err
        }
    }
```



## 关连查询RELS
控制器接收关连关系 rels,  model层查询的时候加入关连参数
```$xslt

    // GetWorkflowBpmnById retrieves WorkflowBpmn by Id. Returns error if
    // Id doesn't exist
    // relations relations data keys
    func GetById(id int, relations ...string) (v Model, err error) {
        gormQuery := db.Conn.Where(id)
    
        //载入关连关系
        for _, rel := range relations {
            gormQuery = gormQuery.Preload(rel)
        }
    
        res := gormQuery.First(&v)
        return v, res.Error
    }

```


## 改造controller层， 加入用户类， 
controller层需继承 github.com/yimishiji/bee/pkg/base/Controller
```$xslt
    import (
        "bpm-api/serviceLogics/HealthChecks"
    
        "github.com/astaxie/beego"
        "github.com/yimishiji/bee/pkg/base"
    )
    
    // SystemController
    type SystemController struct {
        base.Controller
    }
```

获取用户有信息方法, 用户id用string类型记录,如需要int类型，用 strconv.Atoi 转换
```$xslt
    // Post ...
    // @Title Post
    // @Description create WorkflowInstances
    // @Param	body		body 	models.WorkflowInstances	true		"body for WorkflowInstances content"
    // @Success 201 {int} models.WorkflowInstances
    // @Failure 403 body is empty
    // @router / [post]
    func (c *WorkflowInstancesController) Post() {
            CreatedBy := strconv.Atoi(c.User.GetId())
            CompanyId := c.User.GetBusinessID()
            DepartmentId := c.User.GetDepartmentID()
    }
```


 ## filter层实例化，在控制器层实例化
 
 ```$xslt
    // init inputFilter
    func (c *WorkflowInstancesController) Prepare() {
        c.filter = WorkflowInstancesFilter.NewWorkflowInstancesFilter(c.Ctx.Input)
    }
```