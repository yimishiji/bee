// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package apiapp

import (
	"fmt"
	"os"
	path "path/filepath"
	"strings"

	"github.com/yimishiji/bee/cmd/commands"
	"github.com/yimishiji/bee/cmd/commands/version"
	"github.com/yimishiji/bee/generate"
	beeLogger "github.com/yimishiji/bee/logger"
	"github.com/yimishiji/bee/utils"
)

var CmdApiapp = &commands.Command{
	// CustomFlags: true,
	UsageLine: "api [appname]",
	Short:     "Creates a Beego API application",
	Long: `
  The command 'api' creates a Beego API application.

  {{"Example:"|bold}}
      $ bee api [appname] [-tables=""] [-driver=mysql] [-conn="root:@tcp(127.0.0.1:3306)/test"]

  If 'conn' argument is empty, the command will generate an example API application. Otherwise the command
  will connect to your database and generate models based on the existing tables.

  The command 'api' creates a folder named [appname] with the following structure:

	    ├── main.go
	    ├── {{"conf"|foldername}}
	    │     └── app.conf
	    ├── {{"controllers"|foldername}}
	    │     └── object.go
	    │     └── user.go
	    ├── {{"routers"|foldername}}
	    │     └── router.go
	    ├── {{"tests"|foldername}}
	    │     └── default_test.go
	    └── {{"models"|foldername}}
	          └── object.go
	          └── user.go
`,
	PreRun: func(cmd *commands.Command, args []string) { version.ShowShortVersionBanner() },
	Run:    createAPI,
}
var apiconf = `appname = {{.Appname}}
httpport = 8100
runmode = "${DOCKERENV||dev}"
autorender = false
copyrequestbody = true
EnableDocs = true
urlPrefix1 = /{{.Appname}}
DocsPath = /{{.Appname}}/docs

EnableAdmin = true
AdminAddr = ""
AdminPort = 8103

[session]
sessionon = true

[redis]
host = lcoalhost:6391
password = ********
database = 0
prefix = {{.Appname}}_

[db]
host = localhost:3306
user = {{.Appname}}
password = ******
database = {{.Appname}}_db

[smtp]
host = smtp.163.com
prot = 587
user = ***@163.com
password = ****
`
var apiconfLocal = `
[db]
user = {{.Appname}}
password = ******

[redis]
password = ********
database = 0

[smtp]
user = ***@163.com
password = ****
`
var apiMaingo = `package main

import (
	middleWares "{{.Appname}}/pkg/middle-wares"
	_ "{{.Appname}}/routers"
	HealthChecks "{{.Appname}}/service-logics/health-checks"
	"net/http"
	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/astaxie/beego/toolbox"
	"github.com/astaxie/beego/utils"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	//"github.com/yimishiji/bee/pkg/db"
	"github.com/yimishiji/bee/pkg/errors"
)

func init() {

	//连接mysql数据库
	//_, err := db.GetDbConnect()
	//if err != nil {
	//	log.Fatal("mysql Conn :", err)
	//}
	//defer conn.Close()

	//连接redis
	//_, err = db.GetRedisClient()
	//if err != nil {
	//	log.Fatal("redis Conn :", err)
	//}
	//defer client.Close()

	//是否接受panic, 如果设置false 则不接受panic，向上层报错，然后自己接受panic信息；设置true beego会帮你接受panic
	beego.BConfig.RecoverPanic = false
	//是否将错误信息使用html格式打印出来, 设置false不使用html打印
	beego.BConfig.EnableErrorsRender = false

	toolbox.AddHealthCheck("database-mysql", &HealthChecks.DatabaseCheck{})
	toolbox.AddHealthCheck("cache-server-redis", &HealthChecks.RedisCheck{})

}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir[beego.AppConfig.String("DocsPath")] = "swagger"
	}

	localConf := os.Getenv("GOPATH") + "/src/{{.Appname}}/conf/app_local.conf"
	if utils.FileExists(localConf) {
		beego.LoadAppConfig("ini", localConf)
	}

	//跨域请求配置
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	//错误提示
	meihuError := errors.NewMeiHuError()
	beego.ErrorHandler("404", meihuError.PageNotFound)
	beego.ErrorHandler("500", meihuError.ServerInternalError)

	//开启带中间件的web
	beego.RunWithMiddleWares(":"+beego.AppConfig.String("httpport"), MeiHuHandler)
}

//中间件
func MeiHuHandler(handler http.Handler) http.Handler {
	return middleWares.NewMeiHuMiddleWare(handler)
}

`

var apiMainconngo = `package main

import (
	_ "{{.Appname}}/routers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	{{.DriverPkg}}
)

func main() {
	orm.RegisterDataBase("default", "{{.DriverName}}", beego.AppConfig.String("sqlconn"))
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}

`

var apirouter = `// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"{{.Appname}}/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/object",
			beego.NSInclude(
				&controllers.ObjectController{},
			),
		),
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
`

var APIModels = `package ObjectModel

import (
	"errors"
	"strconv"
	"time"
)

var (
	ModelList map[string]*Model
)

type Model struct {
	ObjectId   string
	Score      int64
	PlayerName string
}

func init() {
	ModelList = make(map[string]*Model)
	ModelList["hjkhsbnmn123"] = &Model{"hjkhsbnmn123", 100, "astaxie"}
	ModelList["mjjkxsxsaa23"] = &Model{"mjjkxsxsaa23", 101, "someone"}
}

func AddOne(object Model) (ObjectId string) {
	object.ObjectId = "astaxie" + strconv.FormatInt(time.Now().UnixNano(), 10)
	ModelList[object.ObjectId] = &object
	return object.ObjectId
}

func GetOne(ObjectId string) (object *Model, err error) {
	if v, ok := ModelList[ObjectId]; ok {
		return v, nil
	}
	return nil, errors.New("ObjectId Not Exist")
}

func GetAll() map[string]*Model {
	return ModelList
}

func Update(ObjectId string, Score int64) (err error) {
	if v, ok := ModelList[ObjectId]; ok {
		v.Score = Score
		return nil
	}
	return errors.New("ObjectId Not Exist")
}

func Delete(ObjectId string) {
	delete(ModelList, ObjectId)
}

`

var APIModels2 = `package UserModel

import (
	"errors"
	"strconv"
	"time"
)

var (
	UserList map[string]*Model
)

func init() {
	UserList = make(map[string]*Model)
	u := Model{"user_11111", "astaxie", "11111", Profile{"male", 20, "Singapore", "astaxie@gmail.com"}}
	UserList["user_11111"] = &u
}

type Model struct {
	Id       string
	Username string
	Password string
	Profile  Profile
}

type Profile struct {
	Gender  string
	Age     int
	Address string
	Email   string
}

func Add(u Model) string {
	u.Id = "user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	UserList[u.Id] = &u
	return u.Id
}

func Get(uid string) (u *Model, err error) {
	if u, ok := UserList[uid]; ok {
		return u, nil
	}
	return nil, errors.New("User not exists")
}

func GetAll() map[string]*Model {
	return UserList
}

func Update(uid string, uu *Model) (a *Model, err error) {
	if u, ok := UserList[uid]; ok {
		if uu.Username != "" {
			u.Username = uu.Username
		}
		if uu.Password != "" {
			u.Password = uu.Password
		}
		if uu.Profile.Age != 0 {
			u.Profile.Age = uu.Profile.Age
		}
		if uu.Profile.Address != "" {
			u.Profile.Address = uu.Profile.Address
		}
		if uu.Profile.Gender != "" {
			u.Profile.Gender = uu.Profile.Gender
		}
		if uu.Profile.Email != "" {
			u.Profile.Email = uu.Profile.Email
		}
		return u, nil
	}
	return nil, errors.New("User Not Exist")
}

func Login(username, password string) bool {
	for _, u := range UserList {
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}

func Delete(uid string) {
	delete(UserList, uid)
}
`

var apiControllers = `package controllers

import (
	"{{.Appname}}/models/object"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about object
type ObjectController struct {
	beego.Controller
}

// @Title Create
// @Description create object
// @Param	body		body 	models.Object	true		"The object content"
// @Success 200 {string} ObjectModel.Model.Object.Id
// @Failure 403 body is empty
// @router / [post]
func (o *ObjectController) Post() {
	var ob ObjectModel.Model
	json.Unmarshal(o.Ctx.Input.RequestBody, &ob)
	objectid :=  ObjectModel.AddOne(ob)
	o.Data["json"] = map[string]string{"ObjectId": objectid}
	o.ServeJSON()
}

// @Title Get
// @Description find object by objectid
// @Param	objectId		path 	string	true		"the objectid you want to get"
// @Success 200 {object} ObjectModel.Model
// @Failure 403 :objectId is empty
// @router /:objectId [get]
func (o *ObjectController) Get() {
	objectId := o.Ctx.Input.Param(":objectId")
	if objectId != "" {
		ob, err := ObjectModel.GetOne(objectId)
		if err != nil {
			o.Data["json"] = err.Error()
		} else {
			o.Data["json"] = ob
		}
	}
	o.ServeJSON()
}

// @Title GetAll
// @Description get all objects
// @Success 200 {object} ObjectModel.Model
// @Failure 403 :objectId is empty
// @router / [get]
func (o *ObjectController) GetAll() {
	obs := ObjectModel.GetAll()
	o.Data["json"] = obs
	o.ServeJSON()
}

// @Title Update
// @Description update the object
// @Param	objectId		path 	string	true		"The objectid you want to update"
// @Param	body		body 	models.Object	true		"The body"
// @Success 200 {object} ObjectModel.Model
// @Failure 403 :objectId is empty
// @router /:objectId [put]
func (o *ObjectController) Put() {
	objectId := o.Ctx.Input.Param(":objectId")
	var ob ObjectModel.Model
	json.Unmarshal(o.Ctx.Input.RequestBody, &ob)

	err := ObjectModel.Update(objectId, ob.Score)
	if err != nil {
		o.Data["json"] = err.Error()
	} else {
		o.Data["json"] = "update success!"
	}
	o.ServeJSON()
}

// @Title Delete
// @Description delete the object
// @Param	objectId		path 	string	true		"The objectId you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 objectId is empty
// @router /:objectId [delete]
func (o *ObjectController) Delete() {
	objectId := o.Ctx.Input.Param(":objectId")
	ObjectModel.Delete(objectId)
	o.Data["json"] = "delete success!"
	o.ServeJSON()
}

`
var apiControllers2 = `package controllers

import (
	"{{.Appname}}/models/user"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about Users
type UserController struct {
	beego.Controller
}

// @Title CreateUser
// @Description create users
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {int} UserModel.Model.Id
// @Failure 403 body is empty
// @router / [post]
func (u *UserController) Post() {
	var user UserModel.Model
	json.Unmarshal(u.Ctx.Input.RequestBody, &user)
	uid := UserModel.Add(user)
	u.Data["json"] = map[string]string{"uid": uid}
	u.ServeJSON()
}

// @Title GetAll
// @Description get all Users
// @Success 200 {object} UserModel.Model
// @router / [get]
func (u *UserController) GetAll() {
	users := UserModel.GetAll()
	u.Data["json"] = users
	u.ServeJSON()
}

// @Title Get
// @Description get user by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} UserModel.Model
// @Failure 403 :uid is empty
// @router /:uid [get]
func (u *UserController) Get() {
	uid := u.GetString(":uid")
	if uid != "" {
		user, err := UserModel.Get(uid)
		if err != nil {
			u.Data["json"] = err.Error()
		} else {
			u.Data["json"] = user
		}
	}
	u.ServeJSON()
}

// @Title Update
// @Description update the user
// @Param	uid		path 	string	true		"The uid you want to update"
// @Param	body		body 	UserModel.Model	true		"body for user content"
// @Success 200 {object} UserModel.Model
// @Failure 403 :uid is not int
// @router /:uid [put]
func (u *UserController) Put() {
	uid := u.GetString(":uid")
	if uid != "" {
		var user UserModel.Model
		json.Unmarshal(u.Ctx.Input.RequestBody, &user)
		uu, err := UserModel.Update(uid, &user)
		if err != nil {
			u.Data["json"] = err.Error()
		} else {
			u.Data["json"] = uu
		}
	}
	u.ServeJSON()
}

// @Title Delete
// @Description delete the user
// @Param	uid		path 	string	true		"The uid you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (u *UserController) Delete() {
	uid := u.GetString(":uid")
	UserModel.Delete(uid)
	u.Data["json"] = "delete success!"
	u.ServeJSON()
}

// @Title Login
// @Description Logs user into the system
// @Param	username		query 	string	true		"The username for login"
// @Param	password		query 	string	true		"The password for login"
// @Success 200 {string} login success
// @Failure 403 user not exist
// @router /login [get]
func (u *UserController) Login() {
	username := u.GetString("username")
	password := u.GetString("password")
	if UserModel.Login(username, password) {
		u.Data["json"] = "login success"
	} else {
		u.Data["json"] = "user not exist"
	}
	u.ServeJSON()
}

// @Title logout
// @Description Logs out current logged in user session
// @Success 200 {string} logout success
// @router /logout [get]
func (u *UserController) Logout() {
	u.Data["json"] = "logout success"
	u.ServeJSON()
}

`

var apiTests = `package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"runtime"
	"path/filepath"
	_ "{{.Appname}}/routers"

	"github.com/astaxie/beego"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".." + string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
}

// TestGet is a sample to run an endpoint test
func TestGet(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v1/object", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Test Station Endpoint\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})
}

`

var middleWaresMain = `package middleWares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/yimishiji/bee/pkg/base"
	"{{.Appname}}/service-logics/user"
)

type MeiHuMiddleWare struct {
	h http.Handler
}

func NewMeiHuMiddleWare(h http.Handler) *MeiHuMiddleWare {
	return &MeiHuMiddleWare{
		h: h,
	}
}

//中间件
func (this *MeiHuMiddleWare) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//如果有panic错误，这个会接受一下，并返回到前端中
	defer func() {
		if recInfo := recover(); recInfo != nil {
			var str string
			//panic信息， 把interface 转换string
			str = fmt.Sprint(recInfo) + "; "

			//获取报错的代码行
			for i := 1; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				logs.Critical(fmt.Sprintf("%s:%d", file, line))
				str = str + fmt.Sprintf("%s:%d; ", file, line)
			}
			//数据组装
			resList := &base.Resp{
				TimeTaken: 0,
				Status:    0,
				StatusTxt: str,
				Results: []struct {
				}{},
				Links: "",
			}
			resList.Time = time.Now().Unix()

			newByte, _ := json.Marshal(resList)
			w.Write(newByte)
		}
	}()

	//解析表单
	//r.ParseForm()
	//r.Form.Set("business_id", "1")

	//把url中的横杠（-）去掉
	//r.URL.Path = strings.Replace(r.URL.Path, "-", "", -1)

	//验证操作权限
	if this.checkOperate(r) == false {
		//数据组装
		resList := &base.Resp{
			TimeTaken: 0,
			Status:    base.ApiCode_OAUTH_FAIL,
			StatusTxt: "operate fail",
			Results: []struct {
			}{},
			Links: "",
		}
		resList.Time = time.Now().Unix()
		newByte, _ := json.Marshal(resList)
		w.Write(newByte)
		return
	}

	//向下传递
	this.h.ServeHTTP(w, r)
}

func (this *MeiHuMiddleWare) checkOperate(r *http.Request) bool {
	//判断此url是不是不需要token验证
	OperateKey := getOperateKey(r)
	if isNoTokenUrl(OperateKey) {
		return true
	}

	//预请求跳过权限认证
	if r.Method == "OPTIONS" {
		return true
	}

	//关闭权限验证
	//return true

	//验证用户url访问权限
	return this.VerifyUserOperate(r, OperateKey)
}

//登录用户权限验证p
func (this *MeiHuMiddleWare) VerifyUserOperate(r *http.Request, OerateKey string) bool {
	token := r.Header.Get("Authorization")
	token = strings.Replace(token, "Bearer ", "", 1)
	path := r.URL.Path
	//
	//beego.Info(path, token)
	////API文挡路径
	//if strings.Contains(path, beego.AppConfig.String("DocsPath")) {
	//	beego.Info(path, beego.AppConfig.String("DocsPath"))
	//	return true
	//}

	if UserService.LoginByAccessToken(token) == false {
		return false
	}

	//全部登录用户都可访问列表
	if isAllowAllTokenUrl(OerateKey) {
		return true
	}

	//api文挡请求的跳过授权
	//if strings.HasSuffix(r.Referer(), beego.AppConfig.String("DocsPath")+"/") && r.Method == "GET" {
	//	return true
	//}

	//自定义权限验证
	path = OerateKey
	beego.Info(path)
	if operateList, err := UserService.GetOperateListByAccesstoken(token); err == nil {
		for _, op := range operateList {
			if strings.Trim(op.RightAction, "") == path {
				return true
			}
		}
		beego.Warn("invalid request:" + path)
	} else {
		beego.Warn("invalid request:" + path)
		beego.Warn("get operate list err:", err)
	}
	return false
}

func getOperateKey(r *http.Request) string {
	if r.URL.Path == "" {
		return ""
	}

	prefix := "/{{.Appname}}"
	path := r.URL.Path
	path = strings.Replace(path, prefix, "", 1)
	path = strings.Replace(path, "/v1", "", 1)
	path = strings.TrimRight(path, "/1234567890")
	path = "[" + r.Method + "]" + path
	return path
}
`
var middleWaresMainNotoken = `package middleWares

//不需要token认证的url列表
var noTokenUrlList = map[string]bool{
	"[POST]/user/login":true,
}

//是否不需要token验证的url地址
func isNoTokenUrl(url string) bool {
	if url == "" {
		return false
	}
	//判断key是否存在
	if _, ok := noTokenUrlList[url]; ok {
		return true
	}

	return false
}
`
var middleWaresMainAlltoken = `package middleWares

//不需要token认证的url列表
var allAllowTokenUrlList = map[string]bool{
	"[POST]/object": 		true,
	"[GET]/object":			true,
	"[PUT]/object":			true,
	"[DELETE]/object":  	true,
	"[POST]/user":			true,
	"[GET]/user":			true,
	"[PUT]/user":			true,
	"[DELETE]/user":		true,
	"[GET]/user/logout": 	true,
}

//是否不需要token验证的url地址
func isAllowAllTokenUrl(url string) bool {
	if url == "" {
		return false
	}
	//判断key是否存在
	if _, ok := allAllowTokenUrlList[url]; ok {
		return true
	}

	return false
}
`
var UserServiceTpl = `package UserService

import (
	"github.com/yimishiji/bee/pkg/base"
	"github.com/yimishiji/bee/pkg/db"
	"github.com/astaxie/beego"
	"encoding/json"
)

type RoleRight struct {
	RightAction string
}

//token登录
func LoginByAccessToken(token string) bool {
	if token == "" {
		return false
	}

	user := new(base.User)
	user.AccessToken = token

	userInfoCacheKey := token + ":info"
	if err := db.Redis.Get(userInfoCacheKey).Err(); err == nil {
		return true
	}

	//if userInfo, err := rms.GetUserInfoByToken(token, "{{.Appname}}"); err == nil {
	//	//beego.Info(userInfo)
	//	db.Redis.Set(userInfoCacheKey, userInfo, time.Duration(336)*time.Hour)
	//} else {
	//	beego.Error(err)
	//	return false
	//}

	//if userRight, err := rms.GetRightsByTokenAndPlatformName(token, "{{.Appname}}"); err == nil {
	//	userInfoCacheKey := token + ":right"
	//	db.Redis.Set(userInfoCacheKey, userRight, time.Duration(336)*time.Hour)
	//} else {
	//	return false
	//}

	return true
}

////密码登录后写入session
//func LoginSave(res *RmsApiStructs.UserLoginRes) {
//	token := res.Token.AccessToken
//	userInfoRedisKey := token + ":info"
//	userTokenRedisKey := token + ":token"
//	userRightRedisKey := token + ":right"
//
//	db.Redis.Set(userInfoRedisKey, &res.User, time.Duration(336)*time.Hour)
//	db.Redis.Set(userTokenRedisKey, &res.Token, time.Duration(336)*time.Hour)
//	db.Redis.Set(userRightRedisKey, &res.RoleRight, time.Duration(336)*time.Hour)
//}

// 获取用户的操作权限
func GetOperateListByAccesstoken(token string) (operateList []RoleRight, err error) {

	token = token + ":right"
	data, err := db.Redis.Get(token).Result()
	if err != nil {
		return operateList, err
	}
	err = json.Unmarshal([]byte(data), &operateList)
	if err != nil {
		beego.Error("some error")
	}

	if beego.BConfig.RunMode == "prod" {
		return operateList, err
	}

	//operateList = append(operateList, RoleRight{
	//	RightAction: "[GET]/object",
	//})

	return operateList, nil
}

`

var ServiceDatabaseHealthCheckTpl = `package HealthChecks

import "github.com/yimishiji/bee/pkg/db"

type DatabaseCheck struct {
}

//数据库状态检查
func (c DatabaseCheck) Check() error {
	return db.Conn.DB().Ping()
}
`
var ServiceRedisHealthCheckTpl = `package HealthChecks

import "github.com/yimishiji/bee/pkg/db"

type RedisCheck struct {
}

//redis状态检查
func (c *RedisCheck) Check() error {
	return db.Redis.Ping()
}
`
var beeConfigTpl = `{
	"version": 0,
	"database": {
		"driver": "mysql",
		"conn": "user:password@tcp(localhost:3306)/dbname",
		"prefix": "prefix_"
	},
	"cmd_args": [],
	"enable_reload": true
}
`
var gitIgnoreTpl = `{{.Appname}}*
/vue/
conf/app_local.conf
`
var dockerFileTpl = `FROM golang:1.12.1

# Godep for vendoring
# RUN go get github.com/tools/godep

# Recompile the standard library without CGO
# RUN CGO_ENABLED=0 go install -a std

ENV APP_DIR $GOPATH/src/{{.Appname}}
RUN mkdir -p $APP_DIR

# Set the entrypoint
# ENTRYPOINT ({{.Appname}})
ADD . $APP_DIR

# Compile the binary and statically link
RUN cd $APP_DIR && CGO_ENABLED=0 go build -ldflags '-d -w -s'
# RUN mv $APP_DIR/bpm-api $GOPATH/bin/
# RUN rm -rf $APP_DIR

EXPOSE 8100
`
var dockerIgnoreTpl = `*.zip
*.7z
.git/
vue
{{.Appname}}*
`

func init() {
	CmdApiapp.Flag.Var(&generate.Tables, "tables", "List of table names separated by a comma.")
	CmdApiapp.Flag.Var(&generate.SQLDriver, "driver", "Database driver. Either mysql, postgres or sqlite.")
	CmdApiapp.Flag.Var(&generate.SQLConn, "conn", "Connection string used by the driver to connect to a database instance.")
	commands.AvailableCommands = append(commands.AvailableCommands, CmdApiapp)
}

func createAPI(cmd *commands.Command, args []string) int {
	output := cmd.Out()

	if len(args) < 1 {
		beeLogger.Log.Fatal("Argument [appname] is missing")
	}

	if len(args) > 1 {
		err := cmd.Flag.Parse(args[1:])
		if err != nil {
			beeLogger.Log.Error(err.Error())
		}
	}

	appPath, packPath, err := utils.CheckEnv(args[0])
	appName := path.Base(args[0])
	if err != nil {
		beeLogger.Log.Fatalf("%s", err)
	}
	if generate.SQLDriver == "" {
		generate.SQLDriver = "mysql"
	}

	beeLogger.Log.Info("Creating API...")

	os.MkdirAll(appPath, 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", appPath, "\x1b[0m")
	os.Mkdir(path.Join(appPath, "controllers"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "controllers"), "\x1b[0m")
	os.Mkdir(path.Join(appPath, "tests"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "tests"), "\x1b[0m")

	os.Mkdir(path.Join(appPath, "conf"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "conf"), "\x1b[0m")
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "conf", "app.conf"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "conf", "app.conf"),
		strings.Replace(apiconf, "{{.Appname}}", path.Base(args[0]), -1))
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "conf", "app_local.conf"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "conf", "app_local.conf"),
		strings.Replace(apiconfLocal, "{{.Appname}}", path.Base(args[0]), -1))

	if generate.SQLConn != "" {
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "conf", "app.conf"), "\x1b[0m")
		confContent := strings.Replace(apiconf, "{{.Appname}}", appName, -1)
		confContent = strings.Replace(confContent, "{{.SQLConnStr}}", generate.SQLConn.String(), -1)
		utils.WriteToFile(path.Join(appPath, "conf", "app.conf"), confContent)

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "main.go"), "\x1b[0m")
		mainGoContent := strings.Replace(apiMainconngo, "{{.Appname}}", packPath, -1)
		mainGoContent = strings.Replace(mainGoContent, "{{.DriverName}}", string(generate.SQLDriver), -1)
		if generate.SQLDriver == "mysql" {
			mainGoContent = strings.Replace(mainGoContent, "{{.DriverPkg}}", `_ "github.com/go-sql-driver/mysql"`, -1)
		} else if generate.SQLDriver == "postgres" {
			mainGoContent = strings.Replace(mainGoContent, "{{.DriverPkg}}", `_ "github.com/lib/pq"`, -1)
		}
		utils.WriteToFile(path.Join(appPath, "main.go"),
			strings.Replace(
				mainGoContent,
				"{{.conn}}",
				generate.SQLConn.String(),
				-1,
			),
		)
		beeLogger.Log.Infof("Using '%s' as 'driver'", generate.SQLDriver)
		beeLogger.Log.Infof("Using '%s' as 'conn'", generate.SQLConn)
		beeLogger.Log.Infof("Using '%s' as 'tables'", generate.Tables)
		generate.GenerateAppcode(string(generate.SQLDriver), string(generate.SQLConn), "3", string(generate.Tables), appPath)
	} else {
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "conf", "app.conf"), "\x1b[0m")
		confContent := strings.Replace(apiconf, "{{.Appname}}", appName, -1)
		confContent = strings.Replace(confContent, "{{.SQLConnStr}}", "", -1)
		utils.WriteToFile(path.Join(appPath, "conf", "app.conf"), confContent)

		os.Mkdir(path.Join(appPath, "models"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "models"), "\x1b[0m")
		os.Mkdir(path.Join(appPath, "routers"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "routers")+string(path.Separator), "\x1b[0m")

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "controllers", "object.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "controllers", "object.go"),
			strings.Replace(apiControllers, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "controllers", "user.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "controllers", "user.go"),
			strings.Replace(apiControllers2, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "tests", "default_test.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "tests", "default_test.go"),
			strings.Replace(apiTests, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "routers", "router.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "routers", "router.go"),
			strings.Replace(apirouter, "{{.Appname}}", packPath, -1))

		os.Mkdir(path.Join(appPath, "models", "object"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "models", "object", "model.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "models", "object", "model.go"), APIModels)

		os.Mkdir(path.Join(appPath, "models", "user"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "models", "user", "model.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "models", "user", "model.go"), APIModels2)

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "main.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "main.go"),
			strings.Replace(apiMaingo, "{{.Appname}}", packPath, -1))

		os.Mkdir(path.Join(appPath, "pkg"), 0755)
		os.Mkdir(path.Join(appPath, "pkg", "middle-wares"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "pkg", "middle-wares", "middleware.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "pkg", "middle-wares", "middleware.go"),
			strings.Replace(middleWaresMain, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "pkg", "middle-wares", "allow-all.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "pkg", "middle-wares", "allow-all.go"), middleWaresMainNotoken)

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "pkg", "middle-wares", "allow-token.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "pkg", "middle-wares", "allow-token.go"), middleWaresMainAlltoken)

		os.Mkdir(path.Join(appPath, "service-logics"), 0755)
		os.Mkdir(path.Join(appPath, "service-logics", "user"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "service-logics", "user", "user.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "service-logics", "user", "user.go"),
			strings.Replace(UserServiceTpl, "{{.Appname}}", packPath, -1))

		os.Mkdir(path.Join(appPath, "service-logics", "health-checks"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "service-logics", "health-checks", "database.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "service-logics", "health-checks", "database.go"),
			strings.Replace(ServiceDatabaseHealthCheckTpl, "{{.Appname}}", packPath, -1))
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "service-logics", "health-checks", "redis.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "service-logics", "health-checks", "redis.go"),
			strings.Replace(ServiceRedisHealthCheckTpl, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "bee.json"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "bee.json"),
			strings.Replace(beeConfigTpl, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, ".gitignore"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, ".gitignore"),
			strings.Replace(gitIgnoreTpl, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "Dockerfile"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, "Dockerfile"),
			strings.Replace(dockerFileTpl, "{{.Appname}}", packPath, -1))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, ".dockerignore"), "\x1b[0m")
		utils.WriteToFile(path.Join(appPath, ".dockerignore"),
			strings.Replace(dockerIgnoreTpl, "{{.Appname}}", packPath, -1))

	}
	beeLogger.Log.Success("New API successfully created!")
	return 0
}
