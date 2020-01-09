package etcd

import (
	"encoding/json"
	"fmt"
	"github.com/sunzhiboforever/crontab/common"
	"github.com/sunzhiboforever/crontab/common/class"
	"github.com/sunzhiboforever/crontab/drivers"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strings"
	"time"
)
import "go.etcd.io/etcd/clientv3"

type Etcd struct {
	config      clientv3.Config
	client      *clientv3.Client
	clientKV    clientv3.KV
	clientLease clientv3.Lease
	clientWatch clientv3.Watcher
	//从哪个版本之后开始监听
	revision int64
}

// 初始化 master 的 etcd
func InitEtcdMaster(conf *class.Config) (kv drivers.KvMaster, err error) {
	var etcd Etcd
	etcd.config = clientv3.Config{
		Endpoints:   conf.EtcdEndPoint,
		DialTimeout: time.Duration(conf.EtcdDialTimeout) * time.Millisecond,
	}

	etcd.client, err = clientv3.New(etcd.config)
	etcd.clientKV = clientv3.NewKV(etcd.client)
	etcd.clientLease = clientv3.NewLease(etcd.client)
	return etcd, err
}

// 初始化 worker 的 etcd
func InitEtcdWorker(conf *class.Config) (kv drivers.KvWorker, err error) {
	var etcd Etcd
	etcd.config = clientv3.Config{
		Endpoints:   conf.EtcdEndPoint,
		DialTimeout: time.Duration(conf.EtcdDialTimeout) * time.Millisecond,
	}

	etcd.client, err = clientv3.New(etcd.config)
	etcd.clientKV = clientv3.NewKV(etcd.client)
	etcd.clientLease = clientv3.NewLease(etcd.client)
	etcd.clientWatch = clientv3.NewWatcher(etcd.client)
	return etcd, err
}

// 存入数据
func (e Etcd) SaveJob(key string, value string) (err error) {
	putResponse, err := e.clientKV.Put(context.TODO(), key, value, clientv3.WithPrevKV())
	putResponse = putResponse
	if err != nil {
		return nil
	}
	return
}

// 精准匹配一条数据
func (e Etcd) GetOne(key string) (exists bool, err error) {
	putResponse, err := e.clientKV.Get(context.TODO(), key)
	if err != nil {
		return false, nil
	}
	if len(putResponse.Kvs) != 0 {
		return true, nil
	}
	return false, nil
}

// 获取所有任务
func (e Etcd) GetList() (list map[string]string, err error) {
	var (
		getResponse *clientv3.GetResponse
		res         = make(map[string]string)
	)
	getResponse, err = e.clientKV.Get(context.TODO(), common.DIR_JOB_KEY, clientv3.WithPrefix())
	fmt.Printf("%q\n", getResponse)
	if err != nil {
		return res, err
	}
	for _, mvccpb := range getResponse.Kvs {
		res[string(mvccpb.Key)] = string(mvccpb.Value)
	}
	return res, nil
}

// 监控所有的任务变化，随着worker启动而启动
func (e Etcd) Watch() (eventChan chan *drivers.Event, err error) {
	var (
		getResp       *clientv3.GetResponse
		keyValue      *mvccpb.KeyValue
		watchChan     clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		watchEvents   []*clientv3.Event
		watchEvent    *clientv3.Event
		event         *drivers.Event
		job           *drivers.Job
		eventType     int
		lock          chan int
	)
	// 传输事件channel
	eventChan = make(chan *drivers.Event, 1000)
	// 先把全量数据加载到内存里，再进行监听，这里控制一下先后顺序
	lock = make(chan int, 0)

	// 获取全量任务数据，并把它们视为保存事件
	go func() {
		getResp, err = e.clientKV.Get(context.TODO(), common.DIR_JOB_KEY, clientv3.WithPrefix(), clientv3.WithRev(0))
		if err != nil {
			return
		}
		for _, keyValue = range getResp.Kvs {
			//@todo 这里忽略了json解码失败
			json.Unmarshal(keyValue.Value, &job)
			event = &drivers.Event{
				Type: common.EVENT_TYPE_SAVE,
				Name: strings.TrimPrefix(common.DIR_JOB_KEY, string(keyValue.Key)),
				Job:  job,
			}
			// 推送事件
			eventChan <- event
		}
		lock <- 1
	}()

	// 异步启动常驻协程进行监听
	go func() {
		<-lock
		// 监听目录变化
		watchChan = e.clientWatch.Watch(context.TODO(), common.DIR_JOB_KEY, clientv3.WithPrefix(), clientv3.WithRev(e.revision))
		for watchResponse = range watchChan {
			watchEvents = watchResponse.Events
			for _, watchEvent = range watchEvents {
				switch watchEvent.Type {
				case mvccpb.PUT:
					eventType = common.EVENT_TYPE_SAVE
				case mvccpb.DELETE:
					eventType = common.EVENT_TYPE_DELETE
				}
				//@todo 这里忽略了json解码失败
				json.Unmarshal(watchEvent.Kv.Value, &job)
				// 构建事件
				event = &drivers.Event{
					Type: eventType,
					Name: strings.TrimPrefix(common.DIR_JOB_KEY, string(watchEvent.Kv.Key)),
					Job:  job,
				}
				// 推送事件
				eventChan <- event
			}
		}
	}()
	return
}
