
相对于原生beego, 我们的model层
===



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
rels关连关系 
- Table1Mode
- Table1Mode.table2
- Table1Mode.table3


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
rels关连关系,间接实现表自己关连自己
- Table2Mode
- Table2Mode.Table3Mode
- Table2Mode.Table3Mode.Table2ModeSimple


## 关连关系声明
- model层声明继承的表结构层，及关连的一般model层
```$xslt
type Model struct {
	TableStructs.Flow
	SourceProcess           ProcessModel.Model `json:"source_process" gorm:"foreignkey:ProcessKey;association_foreignkey:SourceRef"`
	TargetProcess           ProcessModel.Model `json:"target_process" gorm:"foreignkey:ProcessKey;association_foreignkey:TargetRef"`
```
- gorm 关连用户参见[http://gorm.io/docs/has_many.html](http://gorm.io/docs/has_many.html)

## 关连查询，model层 GetById方法实现原理
```$xslt
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

## 关连查询，controller层 GetOne方法实现原理
```$xslt
// GetOne ...
// @Title Get One
// @Description get MemberLogController by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Param	rels	query 	string	false		"Many are separated by commas."
// @Success 200 {object} models.MemberLogModel.Model
// @Failure 403 :id is empty
// @router /:id [get]
func (c *MemberLogController) GetOne() {
	id := c.filter.GetId(":id")

	rels := []string{}
	relsStr := strings.Trim(c.Input().Get("rels"), "")
	if relsStr != "" {
		rels = strings.Split(relsStr, ",")
	}

	v, err := UserOriganizationModel.GetById(id, rels...)
	if err != nil {
		c.Data["json"] = c.Resp(base.ApiCode_VALIDATE_ERROR, "not find", err.Error())
	} else {
		c.Data["json"] = c.Resp(base.ApiCode_SUCC, "ok", v)
	}
	c.ServeJSON()
}
```
