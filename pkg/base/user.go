package base

import (
	"encoding/json"

	"strconv"

	"github.com/yimishiji/bee/pkg/db"
)

// 会员基本信息
type User struct {
	Id           string
	AccessToken  string
	RefreshToken string
	UserInfo     *UserInfo
}

//是否是游客
func (u *User) IsGuest() bool {
	return u.Id == "" || u.AccessToken == ""
}

//是否是游客
func (u *User) GetId() string {
	u.Login()
	return u.Id
}

//获取部门id
func (u *User) GetDepartmentID() int {
	u.Login()
	return u.UserInfo.DepartmentID
}

//获取商户id
func (u *User) GetBusinessID() int {
	u.Login()
	return u.UserInfo.BusinessID
}

// 会员附加信息
type UserInfo struct {
	UserID          int    `json:"user_id"`
	LoginAccount    string `json:"login_account"`
	OpenID          string `json:"open_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Mobile          string `json:"mobile"`
	PrefixMobile    string `json:"prefix_mobile"`
	DepartmentID    int    `json:"department_id"`
	PositionID      int    `json:"position_id"`
	JobNumber       int    `json:"job_number"`
	Status          int    `json:"status"`
	Sex             int    `json:"sex"`
	HeadImgURL      string `json:"head_img_url"`
	BusinessID      int    `json:"business_id"`
	IsLoginBusiness int    `json:"is_login_business"`
	Remark          string `json:"remark"`
	Version         int    `json:"version"`
	CreatedBy       string `json:"created_by"`
	UpdatedBy       string `json:"updated_by"`
	UpdatedAt       string `json:"updated_at"`
	CreatedAt       string `json:"created_at"`
	BusinessKey     string `json:"business_key"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	PlatformName string `json:"platform_name"`
	PlatformID   int    `json:"platform_id"`
	UserID       string `json:"user_id"`
}

//登录
func (u *User) Login() {
	if u.AccessToken == "" || u.Id != "" {
		return
	}

	userInfoKey := u.AccessToken + ":info"
	userToken, err := db.Redis.Get(userInfoKey).Result()
	if err != nil {
		return
	}

	userInfo := new(UserInfo)
	err = json.Unmarshal([]byte(userToken), &userInfo)
	if err != nil {
		return
	}

	u.Id = strconv.Itoa(userInfo.UserID)
	u.UserInfo = userInfo
}
