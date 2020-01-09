package master

import (
	"encoding/json"
	"fmt"
	"github.com/sunzhiboforever/crontab/common"
	"github.com/sunzhiboforever/crontab/common/class"
	"github.com/sunzhiboforever/crontab/drivers"
	"net"
	"net/http"
	"strconv"
	"time"
)

// 单例
var InstanceApiServer *apiServer

type apiServer struct {
	httpServer *http.Server
	kvClient   drivers.KvMaster
	mu         map[string]string
}

// 获取任务列表
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		err      error
		list     map[string]string
		response common.Response
		job      common.Job
		jobs     []common.Job
		bytes    []byte
	)
	for {
		list, err = InstanceApiServer.kvClient.GetList()
		fmt.Printf("%q\n", list)
		if err != nil {
			response.ErrNo = 1
			response.ErrMsg = err.Error()
			break
		}
		for _, jobJson := range list {
			fmt.Println(jobJson)
			json.Unmarshal([]byte(jobJson), &job)
			fmt.Println(job)
			jobs = append(jobs, job)
		}
		response.ErrNo = 0
		response.ErrMsg = "success"
		response.Data = jobs
		break
	}
	bytes, err = json.Marshal(response)
	resp.Header().Set("Content-type", "application/json")
	resp.Write(bytes)
}

// 删除一个任务
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		response   common.Response
		postString string
		deleteKey  string
		bytes      []byte
	)
	for {
		err = req.ParseForm()
		if err != nil {
			response.ErrNo = 1
			response.ErrMsg = err.Error()
			break
		}
		postString = req.PostForm.Get("name")
		if _, ok := InstanceApiServer.mu[postString]; ok {
			response.ErrNo = 2
			response.ErrMsg = "system busy"
			break
		}
		if postString == "" {
			response.ErrNo = 3
			response.ErrMsg = "can not receive job name"
			break
		}
		deleteKey = common.DELETE_JOB_KEY + postString
		err = InstanceApiServer.kvClient.SaveJob(deleteKey, "")
		if err != nil {
			response.ErrNo = 6
			response.ErrMsg = err.Error()
			break
		}
		break
	}
	response.ErrNo = 0
	response.ErrMsg = "success"
	delete(InstanceApiServer.mu, postString)
	bytes, err = json.Marshal(response)
	resp.Header().Set("Content-type", "application/json")
	resp.Write(bytes)
}

// 处理保存任务接口
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		postString string
		postData   common.Job
		key        string
		value      string
		response   common.Response
		bytes      []byte
	)
	for {
		err = req.ParseForm()
		if err != nil {
			response.ErrNo = 1
			response.ErrMsg = err.Error()
			break
		}
		postString = req.PostForm.Get("job")
		err = json.Unmarshal([]byte(postString), &postData)
		if err != nil {
			response.ErrNo = 2
			response.ErrMsg = err.Error()
			break
		}
		key = common.DIR_JOB_KEY + postData.Name
		value = postString
		err = InstanceApiServer.kvClient.SaveJob(key, value)
		if err != nil {
			response.ErrNo = 3
			response.ErrMsg = err.Error()
			break
		}
		response.ErrNo = 0
		response.ErrMsg = "success"
		break
	}
	bytes, err = json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		response.ErrMsg = err.Error()
	}
	resp.Header().Set("Content-type", "application/json")
	resp.Write(bytes)
}

func InitApiServer(kvClient drivers.KvMaster) error {
	// 建立路由
	mux := http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)

	// 静态页面路由
	var staticDirString http.Dir
	var staticHandler http.Handler
	staticDirString = http.Dir("../static") //这了和main包不在同一个目录，如有需要要调整一下
	staticHandler = http.FileServer(staticDirString)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

	// 开启监听
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(class.InstanceConfig.ApiPort))
	if err != nil {
		return err
	}

	// 配置服务
	httpServer := &http.Server{
		ReadTimeout:  time.Duration(class.InstanceConfig.ApiReadTimeout) * time.Second,
		WriteTimeout: time.Duration(class.InstanceConfig.ApiWriteTimeout) * time.Second,
		Handler:      mux,
	}
	// 赋单例
	InstanceApiServer = &apiServer{
		httpServer: httpServer,
		kvClient:   kvClient,
		mu:         make(map[string]string),
	}

	go InstanceApiServer.httpServer.Serve(listener)
	return nil
}
