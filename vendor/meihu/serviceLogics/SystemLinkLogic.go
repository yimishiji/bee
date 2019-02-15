package serviceLogics

import (
	"meihu/models"
	"meihu/components/db"
	"fmt"
)

type SystemLinkLogic struct {
}

func (this *SystemLinkLogic) GetSystemLinkList() *[]models.SystemLink {

	systemLink := &[]models.SystemLink{}
	db.Conn.Find(systemLink)
	return systemLink
}

//由于ids传过来是一个切片，如：[]int{1,2,3}，而ids又是一个变长参数,即变长参数都是一个切片，所以格式会是这样 [[1,2,3]], 数组切片，然后再通过...3点打散一下
//那就是[1,2,3] 这种格式了
func (this *SystemLinkLogic) GetSysteLinkListById(ids ...interface{}) *[]models.SystemLink {
	fmt.Println(ids, "ids")
	fmt.Println(ids...)
	systemLink := &[]models.SystemLink{}
	db.Conn.Where("id in (?)", ids...).Find(&systemLink)

	return systemLink
}