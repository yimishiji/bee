package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yimishiji/bee/pkg/base"
)

type MeiHuError struct {
}

func NewMeiHuError() *MeiHuError {
	return &MeiHuError{}
}

//404 找不到页面
func (this *MeiHuError) PageNotFound(w http.ResponseWriter, r *http.Request) {
	resList := base.Resp{
		TimeTaken: 0,
		Status:    base.ApiCode_ILLEGAL_ERROR,
		StatusTxt: "没有找到相关url 404",
		Results:   []struct{}{},
		Links:     "",
	}
	byte, _ := json.Marshal(resList)
	w.Write(byte)
}

//服务器报500错误
func (this *MeiHuError) ServerInternalError(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ServerInternalError")
	resList := base.Resp{
		TimeTaken: 0,
		Status:    base.ApiCode_SYS_ERROR,
		StatusTxt: "server internal error 500",
		Results:   []struct{}{},
		Links:     "",
	}
	byte, _ := json.Marshal(resList)
	w.Write(byte)
}
