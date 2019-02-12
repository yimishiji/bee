package utils

// 接口错误码
type ApiCode int8

// SUCC_11        = 11;// 请求成功,需要弹出confirm窗口却
// SUCC_10        = 10;// 请求成功,需要弹出alert窗口却
// SUCC_2         = 2;// 编号为 2 的成功情况 比如查询成功，但是没有数据
// SUCC           = 1;// 请求成功
// SYS_ERROR      = -1;// 系统错误，一般为操作数据时不成功
// PARAM_ERROR    = -2;// 请求的参数错误或者未通过验证
// VALIDATE_ERROR = -3;// 验证失败
// ILLEGAL_ERROR  = -4;// 非法操作
const (
	ApiCode_SUCC_11        ApiCode = 11
	ApiCode_SUCC_10        ApiCode = 10
	ApiCode_SUCC_2         ApiCode = 2
	ApiCode_SUCC           ApiCode = 1 // 请求成功
	ApiCode_SYS_ERROR      ApiCode = -1
	ApiCode_PARAM_ERROR    ApiCode = -2
	ApiCode_VALIDATE_ERROR ApiCode = -3
	ApiCode_ILLEGAL_ERROR  ApiCode = -4
)

// Resp
type Resp struct {
	TimeTaken float64     `json:"time_token"`
	Status    ApiCode     `json:"status"`
	StatusTxt string      `json:"status_txt"`
	Results   interface{} `json:"result"`
	Links     string      `json:"links"`
	Time      int64       `json:"time"`
}
