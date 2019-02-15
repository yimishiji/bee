package controllers

import (
	"meihu/filters"
	"meihu/components/db"
	"fmt"
	"strings"
	"time"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"meihu/serviceLogics"
)
var (
	userFilter *filters.UserFilter
)

type UserController struct {
	BaseController
}

func (this *UserController) Prepare() {
	userFilter = filters.NewUserFilter(this.Ctx.Request)
}

func (this *UserController) GetList() {
	userList, err := userFilter.GetUserList()
	if err != nil {
		this.SetReponse(nil, err.Error())
		return
	}

	this.SetReponse(userList)
}

func (this *UserController) AddUser() {
	userFilter.AddUser()
}

func (this *UserController) EditUser() {

}


func (this *UserController) DeleteUser() {

}

func (this *UserController) Test() {
	start := time.Now()
	t := &serviceLogics.Test{}
	u := new(serviceLogics.UserLogic)
	u.Test(t)

	end := time.Since(start).Seconds()
	t.GoTime = end


	fmt.Println(a)

	this.SetReponse(t)

}



