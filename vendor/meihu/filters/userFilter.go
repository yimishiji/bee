package filters

import (
	"meihu/serviceLogics"
	"meihu/models"
	"net/http"
	"github.com/astaxie/beego/validation"
	"errors"
	"regexp"
	"fmt"
)

type UserFilter struct {
	Req *http.Request
}

func NewUserFilter(r *http.Request) *UserFilter {
	return &UserFilter{
		Req: r,
	}
}

//数据过滤
func (this *UserFilter) GetUserList() (userList *[]models.Users, err error) {
	pageSize := "10"
	if this.Req.FormValue("page_size") != "" {
		pageSize = this.Req.FormValue("page_size")
	}

	currentPage := "1"
	if this.Req.FormValue("current_page") != "" {
		currentPage = this.Req.FormValue("current_page")
	}

	businessId := this.Req.FormValue("business_id")

	valid := validation.Validation{}
	valid.Numeric(pageSize, "page_size").Message("pageSize必须是数字")
	valid.Numeric(currentPage, "current_page").Message("currentPage必须是数字")
	if valid.HasErrors() {
		err = errors.New(valid.Errors[0].String())
		return nil, err
	}

	userLogic := new(serviceLogics.UserLogic)
	userList = userLogic.GetUserList(pageSize, currentPage, businessId)

	return userList, nil
}

func (this *UserFilter) AddUser()  {
	valid := validation.Validation{}

	userName := this.Req.FormValue("user_name")
	//正则校验
	valid.Match(userName, regexp.MustCompile(`^[\w-_]{4,20}$`), "user_name").Message("用户名格式不合法")
	if valid.HasErrors() {
		err := errors.New(valid.Errors[0].String())
		fmt.Println(err, "aaa")
	}

	//name := this.Req.FormValue("name")
	//enName := this.Req.FormValue("en_name")
	//codeName := this.Req.FormValue("code_name")
	//password := this.Req.FormValue("password")
	//mobile := this.Req.FormValue("mobile")
	//sex := this.Req.FormValue("sex")
	//email := this.Req.FormValue("email")
	//jobNumber := this.Req.FormValue("job_number")
	//positionId := this.Req.FormValue("position_id")
	//depId := this.Req.FormValue("dep_id")
	//schemeId := this.Req.FormValue("scheme_id")
	//salary := this.Req.FormValue("salary")
	//entryTime := this.Req.FormValue("entry_time")
	//idNumber := this.Req.FormValue("id_number")
	//dateOfBirth := this.Req.FormValue("date_of_birth")
	//province := this.Req.FormValue("province")
	//city := this.Req.FormValue("city")
	//district := this.Req.FormValue("district")
	//address := this.Req.FormValue("address")
	//localAddress := this.Req.FormValue("local_address")
	//contractExpirationTime := this.Req.FormValue("contract_expiration_time")
	//healthCardExpirationTime := this.Req.FormValue("health_card_expiration_time")


}