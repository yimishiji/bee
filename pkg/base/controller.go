// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package base

import (
	"errors"
	"time"

	"strings"

	"github.com/astaxie/beego"
)

// BpmWorkflowsController operations for BpmWorkflows
type Controller struct {
	beego.Controller
}

//输出格式统一处理
func (c *Controller) Resp(appCode ApiCode, msg string, data ...interface{}) *Resp {
	resp := new(Resp)
	resp.Results = data
	resp.Status = appCode
	resp.StatusTxt = msg
	resp.TimeTaken = c.Ctx.ResponseWriter.Elapsed.Seconds()
	resp.Time = time.Now().Unix()
	return resp
}

func (c *Controller) GetPagePublicParams() (fields []string, sortby []string, query map[string]string, limit int64, offset int64, err error) {
	var field []string
	var sort []string
	var orders []string
	var querys = make(map[string]string)
	var limits int64 = 10
	var offsets int64 = 0
	var sortFields []string

	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		field = strings.Split(v, ",")
	}

	// limit: 10 (default is 10)
	if v, err := c.GetInt64("limit"); err == nil {
		limits = v
	}
	// offset: 0 (default is 0)
	if v, err := c.GetInt64("offset"); err == nil {
		offsets = v
	}
	// sortby: col1,col2
	if v := c.GetString("sortby"); v != "" {
		sort = strings.Split(v, ",")
	}
	// order: desc,asc
	if v := c.GetString("order"); v != "" {
		orders = strings.Split(v, ",")
	}
	// query: k:v,k:v
	if v := c.GetString("query"); v != "" {
		for _, cond := range strings.Split(v, ",") {
			kv := strings.SplitN(cond, ":", 2)
			if len(kv) != 2 {
				return field, sortFields, querys, limits, offsets, errors.New("Error: invalid query key/value pair")
			}
			k, v := kv[0], kv[1]
			querys[k] = v
		}
	}

	// order by:
	if len(sort) != 0 {
		if len(sort) == len(orders) {
			// 1) for each sort field, there is an associated order
			for i, v := range sort {
				orderby := ""
				if orders[i] == "desc" {
					orderby = "-" + v
				} else if orders[i] == "asc" {
					orderby = v
				} else {
					return field, sortFields, querys, limits, offsets, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sort) != len(orders) && len(orders) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sort {
				orderby := ""
				if orders[0] == "desc" {
					orderby = "-" + v
				} else if orders[0] == "asc" {
					orderby = v
				} else {
					return field, sortFields, querys, limits, offsets, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sort) != len(orders) && len(orders) != 1 {
			return field, sortFields, querys, limits, offsets, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(orders) != 0 {
			return field, sortFields, querys, limits, offsets, errors.New("Error: unused 'order' fields")
		}
	}

	return field, sortFields, querys, limits, offsets, nil
}
