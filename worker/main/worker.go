package main

import (
	"flag"
	"fmt"
	"github.com/sunzhiboforever/crontab/common/class"
	"github.com/sunzhiboforever/crontab/drivers/etcd"
	"github.com/sunzhiboforever/crontab/worker"
	"runtime"
	"sync"
)

func initEnv()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var (
	err error
	confFile string //命令行参数
)

func initArgs()  {
	flag.StringVar(&confFile, "config", "./worker.json", "配置文件路径")
	flag.Parse()
}

func main()  {
	// 初始化系统线程
	initEnv()

	// 初始化命令行参数
	initArgs()

	// 初始化配置文件
	err = class.InitConfig(confFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 初始化配置系统
	//@todo 根据不同的配置，选择不同的系统
	kv, err := etcd.InitEtcdWorker(class.InstanceConfig)
	if err != nil {
		fmt.Println(err)
	}
	worker.InitSchedule(kv)

	w := sync.WaitGroup{}
	w.Add(1)
	w.Wait()
}
