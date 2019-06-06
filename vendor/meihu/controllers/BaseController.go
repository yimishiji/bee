package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
)

type ReponseList struct {
	Code int	`json:"code"`
	Message string	`json:"message"`
	Data interface{}	`json:"data"`
}

type BaseController struct {
	beego.Controller
}

func (this *BaseController) Prepare() {
	fmt.Println("prepare")
	fmt.Println(this.GetControllerAndAction())
}

//code int, message string, data interface{}
/**
params1 data
params2 message
params3 code
 */
func (this *BaseController) SetReponse(params ...interface{}) {
	var code int = 1
	var message string
	var data interface{}

	len := len(params)
	if len > 0 {
		if params[0] != nil {
			data = params[0]
		} else {
			data = []struct {}{}
		}
	} else {
		data = []struct {}{}
	}

	if len > 1 {
		message = params[1].(string)
		code = 0;
	}

	if len > 2 {
		code = params[2].(int)
	}

	responseList := ReponseList{
		Code: code,
		Message: message,
		Data:data,
	}

	this.Data["json"] = responseList
	this.ServeJSON()
}

func (this *BaseController) Redirect(url string) {
	this.Ctx.Redirect(302, url)
}

func (this *BaseController) Finish() {
	fmt.Println("finish ....")
}