package class

import "flag"

var (
	confFile string //命令行参数
)

func InitArgs() {
	flag.StringVar(&confFile, "config", "./master.json", "配置文件路径")
	flag.Parse()
}

