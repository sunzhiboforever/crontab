package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		kv     clientv3.KV
		//getResp   *clientv3.GetResponse
		//putResp   *clientv3.PutResponse
		delResp   *clientv3.DeleteResponse
		watcher   clientv3.Watcher
		watchChan clientv3.WatchChan
		lease         clientv3.Lease
		leaseId       clientv3.LeaseID
		leaseResp     *clientv3.LeaseGrantResponse
		leaseRespChan <-chan *clientv3.LeaseKeepAliveResponse
		ctx       context.Context
		cancelFun context.CancelFunc
		//op      clientv3.Op
		txn     clientv3.Txn
		txnResp *clientv3.TxnResponse
	)

	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
	}
	watcher = clientv3.NewWatcher(client)
	watchChan = watcher.Watch(context.TODO(), "test1")
	go func(watchChan clientv3.WatchChan) {
		for WatchResponse := range watchChan {
			for _, event := range WatchResponse.Events {
				switch event.Type {
				case mvccpb.PUT:
					fmt.Println("进行了put操作", event.Kv.ModRevision)
				case mvccpb.DELETE:
					fmt.Println("进行了delete操作", event.Kv.ModRevision)

				}
			}
		}
	}(watchChan)
	kv = clientv3.NewKV(client)

	for i := 1; i < 10; i++ {
		//fmt.Println("放入value：", i)
		if _, err = kv.Put(context.TODO(), "test1", string(i), clientv3.WithPrevKV()); err != nil {
			fmt.Println(err)
			break
		}
		//fmt.Println("上一个value是：", putResp.PrevKv.Value)
		if _, err = kv.Get(context.TODO(), "test1"); err != nil {
			fmt.Println(err)
		}
		//fmt.Println("现在的value是：", getResp.Kvs[0].Value)
		//fmt.Println()
	}
	if delResp, err = kv.Delete(context.TODO(), "test1", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
		return
	}
	if delResp.Deleted > 0 {
		fmt.Println("删除成功，之前的值为：", delResp.PrevKvs[0].Value)
	}



	//设置超时 10 秒
	ctx, cancelFun = context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFun()
	//申请租约租约
	lease = clientv3.NewLease(client)
	if leaseResp, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println(err)
		return
	}
	leaseId = leaseResp.ID
	defer lease.Revoke(context.TODO(), leaseId)

	// 创建事务
	txn = kv.Txn(context.TODO())
	txn.If(clientv3.Compare(clientv3.CreateRevision("test"), "=", 0)).
		Then(clientv3.OpPut("test", "xxx", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("test"))
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}
	// 如果事务没有执行成功，说明锁被抢占
	if !txnResp.Succeeded {
		fmt.Println("锁被占用了", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	//处理任务并且续租
	fmt.Println("开始处理任务")
	if leaseRespChan, err = lease.KeepAlive(ctx, leaseId); err != nil {
		return
	}
	go func(leaseRespChan <-chan *clientv3.LeaseKeepAliveResponse) {
		for leaseResp := range leaseRespChan {
			fmt.Println("租约id：", leaseResp.ID, "续租时间 5 秒中")
		}
	}(leaseRespChan)
	time.Sleep(10 * time.Second)

}
