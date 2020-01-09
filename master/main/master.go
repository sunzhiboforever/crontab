package main

import (
	"flag"
	"fmt"
	"github.com/sunzhiboforever/crontab/common/class"
	"github.com/sunzhiboforever/crontab/drivers/etcd"
	"github.com/sunzhiboforever/crontab/master"
	"runtime"
	"sync"
)

func initEnv()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var (
	err error
	confFile string //配置文件路径
)

func initArgs()  {
	flag.StringVar(&confFile, "config", "./master.json", "配置文件路径")
	flag.Parse()
}

func main()  {
	// 初始化线程
	initEnv()

	// 初始化命令行参数
	initArgs()

	// 初始化配置文件
	err = class.InitConfig(confFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 初始化etcd
	//@todo 根据不同的配置，选择不同的系统
	kv, err := etcd.InitEtcdMaster(class.InstanceConfig)
	if err != nil {
		fmt.Println(err)
	}
	// 启动 web api
	err = master.InitApiServer(kv)
	if err != nil {
		fmt.Println(err)
	}

	w := sync.WaitGroup{}
	w.Add(1)
	w.Wait()
}
