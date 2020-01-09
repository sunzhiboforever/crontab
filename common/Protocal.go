package common

// 后台插入任务
type Job struct {
	Name string	`json:"name"`
	Command string	`json:"command"`
	CronExpr string	`json:"cron_expr"`
}

type Response struct {
	// 0：木有问题
	ErrNo int	`json:"err_no"`
	ErrMsg string	`json:"err_msg"`
	Data interface{}	`json:"data"`
}
