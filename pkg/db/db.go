package db

import (
	"fmt"

	"reflect"
	"strconv"
	"strings"

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

//过滤条件
func NewGormQuery(query map[string]string) *gorm.DB {
	gorm := Conn
	//过滤条件
	for k, v := range query {
		if strings.Contains(k, "-isnull") {
			k = strings.Replace(k, "-isnull", "", 1)
			gorm = gorm.Where(k + " isnull")
		} else if strings.HasPrefix(v, ">=") {
			v = strings.Replace(v, ">=", "", 1)
			gorm = gorm.Where(k+" >= ?", v)
		} else if strings.HasPrefix(v, "<=") {
			v = strings.Replace(v, "<=", "", 1)
			gorm = gorm.Where(k+" <= ?", v)
		} else if strings.HasPrefix(v, ">") {
			v = strings.Replace(v, ">", "", 1)
			gorm = gorm.Where(k+" > ?", v)
		} else if strings.HasPrefix(v, "<") {
			v = strings.Replace(v, "<", "", 1)
			gorm = gorm.Where(k+" < ?", v)
		} else if strings.HasPrefix(v, "!=") {
			v = strings.Replace(v, "!=", "", 1)
			gorm = gorm.Where(k+" != ?", v)
		} else if strings.HasPrefix(v, "<>") {
			v = strings.Replace(v, "<>", "", 1)
			gorm = gorm.Where(k+" <> ?", v)
		} else if strings.HasPrefix(v, "like-") {
			v = strings.Replace(v, "like-", "", 1)
			gorm = gorm.Where(k+" LIKE ?", "%"+v+"%")
		} else if strings.HasPrefix(v, "between-") {
			v = strings.Replace(v, "between-", "", 1)
			ranges := strings.SplitN(v, "-", 2)
			if len(ranges) == 2 {
				star, _ := strconv.Atoi(ranges[0])
				end, _ := strconv.Atoi(ranges[1])
				gorm = gorm.Where(k+" BETWEEN ? AND ?", star, end)
			}
		} else if strings.HasPrefix(v, "in-") {
			v = strings.Replace(v, "in-", "", 1)
			ranges := strings.SplitN(v, "-", -1)
			gorm = gorm.Where(k+" in (?)", ranges)
		} else {
			gorm = gorm.Where(k+" = ?", v)
		}
	}

	return gorm
}

//过滤字段，实现select功能
func SelectField(l []interface{}, fields []string) (ml []interface{}) {
	if len(fields) == 0 {
		return l
	} else {
		// trim unused fields
		for _, v := range l {
			m := make(map[string]interface{})
			val := reflect.ValueOf(v)
			for _, fname := range fields {
				m[fname] = val.FieldByName(fname).Interface()
			}
			ml = append(ml, m)
		}
	}
	return ml
}
