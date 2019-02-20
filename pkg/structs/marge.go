package structs

import (
	"encoding/json"
)

//合并结构体数据
func StructMarge(c1 interface{}, c2 interface{}) {
	json2, _ := json.Marshal(c2)
	json.Unmarshal(json2, &c1)
}
