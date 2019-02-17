package base

import (
	"encoding/json"

	"github.com/yimishiji/bee/pkg/db"
)

// 会员基本信息
type User struct {
	Id           string
	AccessToken  string
	RefreshToken string
	UserInfo     UserInfo
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

// 会员附加信息
type UserInfo struct {
	UserID       string `json:"user_id"`
	LoginAccount string `json:"login_account"`
	OpenID       string `json:"open_id"`
	Name         string `json:"name"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	Mobile       string `json:"mobile"`
	JobNumber    string `json:"job_number"`
	Status       string `json:"status"`
	Sex          string `json:"sex"`
	HeadImgURL   string `json:"head_img_url"`
	BusinessID   string `json:"business_id"`
	Remark       string `json:"remark"`
	Version      string `json:"version"`
	CreatedBy    string `json:"created_by"`
	UpdatedBy    string `json:"updated_by"`
	UpdatedAt    string `json:"updated_at"`
	CreatedAt    string `json:"created_at"`
	BusinessKey  string `json:"business_key"`
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

	userInfoKey := u.AccessToken + ":token"
	userToken, err := db.Redis.Get(userInfoKey).Result()
	if err != nil {
		return
	}

	token := new(Token)
	err = json.Unmarshal([]byte(userToken), &token)
	if err != nil {
		return
	}

	u.RefreshToken = token.RefreshToken
	u.Id = token.UserID
}
