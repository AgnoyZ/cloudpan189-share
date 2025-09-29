package autoingestlog

// ListRequest 自动挂载日志列表查询请求
type ListRequest struct {
	PlanId      int64
	Level       string
	CurrentPage int
	PageSize    int
}
