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
	"github.com/yimishiji/bee/utils"
)

// BpmWorkflowsController operations for BpmWorkflows
type Controller struct {
	beego.Controller
}

//输出格式统一处理
func (c *Controller) Resp(appCode utils.ApiCode, msg string, data ...interface{}) *utils.Resp {
	resp := new(utils.Resp)
	resp.Results = data
	resp.Status = appCode
	resp.StatusTxt = msg
	resp.TimeTaken = c.Ctx.ResponseWriter.Elapsed.Seconds()
	resp.Time = time.Now().Unix()
	return resp
}

func (c *Controller) GetPagePublicParams() (fields []string, sortby []string, order []string, query map[string]string, limit int64, offset int64, err error) {
	var field []string
	var sort []string
	var orders []string
	var querys = make(map[string]string)
	var limits int64 = 10
	var offsets int64

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
				return field, sort, orders, querys, limits, offsets, errors.New("Error: invalid query key/value pair")
			}
			k, v := kv[0], kv[1]
			querys[k] = v
		}
	}

	return field, sort, orders, querys, limits, offsets, err
}
