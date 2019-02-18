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
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

type beeController struct {
	beego.Controller
}

// BpmWorkflowsController operations for BpmWorkflows
type Controller struct {
	beeController
	accessToken string
	User        User
}

// Init generates default values of controller operations.
func (c *Controller) Init(ctx *context.Context, controllerName, actionName string, app interface{}) {
	c.beeController.Init(ctx, controllerName, actionName, app)

	//初始化用户token
	token := c.Ctx.Input.Header("Authorization")
	token = strings.Replace(token, "Bearer ", "", 1)
	c.User.AccessToken = token
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
