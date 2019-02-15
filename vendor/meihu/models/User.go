package models

type Users struct {
	UserId	uint `json:"user_id"`
	UserName string `json:"user_name"`
	Name string `json:"name"`
	EnName string `json:"en_name"`
	CodeName uint `json:"code_name"`
	BusinessId uint `json:"business_id"`
	Status uint8 `json:"status"`
	Password string `json:"password"`
	Mobile string `json:"mobile"`
	Sex uint8 `json:"sex"`
	Email string `json:"email"`
	HeadImage string `json:"head_image"`
	JobNumber uint `json:"job_number"`
	OpenId string `json:"open_id"`
	PositionId uint `json:"position_id"`
	DepId uint `json:"dep_id"`
	SchemeId uint `json:"scheme_id"`
	Salary	float32 `json:"salary"`
	EntryTime string `json:"entry_time"`
	IdNumber string `json:"id_number"`
	DateOfBirth	string `json:"date_of_birth"`
	Province string `json:"province"`
	City string `json:"city"`
	District string `json:"district"`
	Address string `json:"address"`
	LocalAddress string `json:"local_address"`
	ContractExpirationTime string `json:"contract_expiration_time"`
	HealthCardExpirationTime string `json:"health_card_expiration_time"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}