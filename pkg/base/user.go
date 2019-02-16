package base

var UserInstance = new(User)

// 会员基本信息
type User struct {
	Id          string
	AccessToken string
	UserInfo    UserInfo
}

//是否是游客
func (u *User) isGuest() bool {
	return u.Id == "" || u.AccessToken == ""
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
