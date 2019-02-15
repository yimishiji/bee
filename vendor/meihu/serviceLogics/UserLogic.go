package serviceLogics

import (
	"meihu/models"
	"meihu/components/db"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

type UserLogic struct {
	CommonLogic
}

type Test struct {
	TimeToken string `json:"time_token"`
	Status int `json:"status"`
	StatusTxt string `json:"status_txt"`
	Results []string `json:"results"`
	Time int `json:"time"`
	GoTime float64 `json:"go_time"`
}

func (this *UserLogic) GetUserList(pageSize string, currentPage string, businessId string) *[]models.Users{
	userList := &[]models.Users{}

	pSize, _ := strconv.ParseInt(pageSize, 10, 32)
	cPage, _ := strconv.ParseInt(currentPage, 10, 32)

	offset :=  pSize * (cPage - 1)
	db.Conn.Where("business_id = ?", businessId).Limit(pSize).Offset(offset).Find(userList)

	return userList
}

func (this *UserLogic) AddUser(userList *models.Users)  {

	db.Conn.Create(userList)
}

func (this *UserLogic) Test(t *Test) {

	tmp := make(chan int, 1)
	go func() {

		resp, err := http.Get("https://ms.yimishiji.com/rms-api/admin/oauth2")
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()

		buffer, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(buffer, t)
		tmp <- 1
	}()


	<- tmp
}