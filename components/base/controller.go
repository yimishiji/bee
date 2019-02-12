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
	"time"

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
