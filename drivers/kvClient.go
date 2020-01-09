package drivers

// master的面向后端web页面的后端接口接口定义
type KvMaster interface {
	// 后台新建/修改任务接口
	SaveJob(key, value string) (err error)

	// 后台判断任务是否存在的接口
	GetOne(key string) (exists bool, err error)

	// 后台获取任务列表接口
	GetList() (list map[string]string, err error)
}

// 定义一个任务类型
type Job struct {
	//任务名称
	Name string

	//任务命令定义
	Command string

	//任务的cron表达式
	CronExpr string
}

// 定义一个任务变动事件
type Event struct {
	//事件类型   0:新建事件 1:删除事件 2: 杀死事件
	Type int

	//任务名称
	Name string

	//具体的任务
	Job *Job
}

// worker的kv存储接口定义
type KvWorker interface {
	// 监控所有的任务，如果有变动，返回一个事件
	Watch() (eventChan chan *Event, err error)
}
