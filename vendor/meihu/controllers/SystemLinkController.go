package controllers

import (
	"fmt"
	"meihu/serviceLogics"
)

type SystemLinkController struct {
	BaseController
}

func (u *SystemLinkController) GetList() {
	systemLinkLogic := new(serviceLogics.SystemLinkLogic)
	systemLink := systemLinkLogic.GetSystemLinkList()
	fmt.Println(systemLink)

	u.SetReponse(systemLink)
}

func (u *SystemLinkController) GetLinkById() {

	systemLinkLogic := new(serviceLogics.SystemLinkLogic)
	systemLink := systemLinkLogic.GetSysteLinkListById([]int{6})
	fmt.Println(u.Ctx.Request)

	u.SetReponse(systemLink)
}
