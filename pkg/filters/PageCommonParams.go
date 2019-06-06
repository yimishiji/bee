package filters

//公类页公共参数结构
type PageCommonParams struct {
	Field      []string
	Sort       []string
	Orders     []string
	Querys     map[string]string
	Limits     int64
	Offsets    int64
	SortFields []string
	Rels       []string
}
