package filters

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego/context"
)

type InputFilter struct {
	Input *context.BeegoInput
}

//获取分类页公共请求参数
func (c *InputFilter) GetPagePublicParams() (p *PageCommonParams, err error) {

	var params = &PageCommonParams{
		Limits:  10,
		Offsets: 0,
	}

	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		params.Field = strings.Split(v, ",")
	}

	// limit: 10 (default is 10)
	if v, err := c.GetInt64("limit"); err == nil {
		params.Limits = v
	}
	// offset: 0 (default is 0)
	if v, err := c.GetInt64("offset"); err == nil {
		params.Offsets = v
	} else if page, err := c.GetInt64("page"); err == nil {
		params.Offsets = (page - 1) * params.Limits
	}

	// sortby: col1,col2
	if v := c.GetString("sortby"); v != "" {
		params.Sort = strings.Split(v, ",")
	}
	// order: desc,asc
	if v := c.GetString("order"); v != "" {
		params.Orders = strings.Split(v, ",")
	}
	// query: k:v,k:v

	var query = make(map[string]string)
	if v := c.GetString("query"); v != "" {
		fmt.Println(v)

		for _, cond := range strings.Split(v, ",") {
			kv := strings.SplitN(cond, ":", 2)
			if len(kv) != 2 {
				return params, errors.New("Error: invalid query key/value pair")
			}
			k, v := kv[0], kv[1]
			query[k] = v
		}
	}
	params.Querys = query

	//relations data
	if v := c.GetString("rels"); v != "" {
		for _, rel := range strings.Split(v, ",") {
			if rel = strings.Trim(rel, ""); rel != "" {
				params.Rels = append(params.Rels, rel)
			}
		}
	}

	// order by:
	if len(params.Sort) != 0 {
		if len(params.Sort) == len(params.Orders) {
			// 1) for each sort field, there is an associated order
			for i, v := range params.Sort {
				orderby := ""
				if params.Orders[i] == "desc" {
					orderby = "-" + v
				} else if params.Orders[i] == "asc" {
					orderby = v
				} else {
					return params, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				params.SortFields = append(params.SortFields, orderby)
			}
		} else if len(params.Sort) != len(params.Orders) && len(params.Orders) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range params.Sort {
				orderby := ""
				if params.Orders[0] == "desc" {
					orderby = "-" + v
				} else if params.Orders[0] == "asc" {
					orderby = v
				} else {
					return params, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				params.SortFields = append(params.SortFields, orderby)
			}
		} else if len(params.Sort) != len(params.Orders) && len(params.Orders) != 1 {
			return params, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(params.Orders) != 0 {
			return params, errors.New("Error: unused 'order' fields")
		}
	}

	return params, nil
}

// GetString returns the input value by key string or the default value while it's present and input is blank
func (c *InputFilter) GetString(key string, def ...string) string {
	if v := c.Input.Query(key); v != "" {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetInt64 returns input value as int64 or the default value while it's present and input is blank.
func (c *InputFilter) GetInt64(key string, def ...int64) (int64, error) {
	strv := c.Input.Query(key)
	if len(strv) == 0 && len(def) > 0 {
		return def[0], nil
	}
	return strconv.ParseInt(strv, 10, 64)
}

//获取paramsKeyId
func (c *InputFilter) GetId(key string) int {
	idStr := c.Input.Param(key)
	id, _ := strconv.Atoi(idStr)
	return id
}

//判断字符串是否在数组中
func InStingArr(str string, needArr []string) bool {
	for _, v := range needArr {
		if str == v {
			return true
		}
	}
	return false
}
