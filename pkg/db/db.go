package db

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/jinzhu/gorm"
)

var (
	Conn *gorm.DB
)

func GetDbConnect() (*gorm.DB, error) {
	dbConStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true", beego.AppConfig.String("db::user"), beego.AppConfig.String("db::password"), beego.AppConfig.String("db::host"), beego.AppConfig.String("db::database"))
	db, err := gorm.Open("mysql", dbConStr)
	Conn = db
	return db, err
}
