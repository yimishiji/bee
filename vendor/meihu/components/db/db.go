package db

import (
	"github.com/jinzhu/gorm"
	"fmt"
	"github.com/astaxie/beego"
)

var (
	Conn *gorm.DB
)

func GetDbConnect() (*gorm.DB, error) {
	dbConStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", beego.AppConfig.String("db::user"), beego.AppConfig.String("db::password"), beego.AppConfig.String("db::host"), beego.AppConfig.String("db::database"))
	db, err := gorm.Open("mysql", dbConStr)
	Conn = db
	return db, err
}
