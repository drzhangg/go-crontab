package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

/*
	watch 监听目录变化
 */
func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err error
		kv clientv3.KV
		watcher clientv3.Watcher
		getResp *clientv3.GetResponse
		watchStartRevision int64
		watchRespChan <- chan clientv3.WatchResponse
		watchResp clientv3.WatchResponse
		event *clientv3.Event
	)

	config = clientv3.Config{
		Endpoints:[]string{"47.99.240.52:2379"},
		DialTimeout:5 * time.Second,
	}

	//创建一个客户端
	if client,err  = clientv3.New(config);err != nil{
		fmt.Println(err)
		return
	}

	//创建一个kv
	kv = clientv3.NewKV(client)

	go func() {
		for{
			kv.Put(context.TODO(),"/cron/jobs/job7","i am job7")

			kv.Delete(context.TODO(),"/cron/jobs/job7")

			time.Sleep(time.Second)
		}
	}()

	//先GET到当前的值，并监听后续变化
	if getResp,err = kv.Get(context.TODO(),"/cron/jobs/job7");err != nil {
		fmt.Println(err)
		return
	}

	//现在key是空的
	if len(getResp.Kvs) != 0 {
		fmt.Println("当前值:",string(getResp.Kvs[0].Value))
	}


	//当前etcd集群事务ID，单调递增的
	watchStartRevision = getResp.Header.Revision + 1

	//创建一个watcher
	watcher = clientv3.NewWatcher(client)

	//启动监听
	fmt.Println("从该版本向后监听：",watchStartRevision)

	ctx,cancelFun := context.WithCancel(context.TODO())
	time.AfterFunc(5 * time.Second, func() {
		cancelFun()
	})

	watchRespChan = watcher.Watch(ctx,"/cron/jobs/job7",clientv3.WithRev(watchStartRevision))


	//处理kv变化事件
	for watchResp = range watchRespChan{
		for _,event = range watchResp.Events{
			switch event.Type {
			case mvccpb.PUT:
				fmt.Println("修改为：",string(event.Kv.Value),	"Revision:",event.Kv.CreateRevision,event.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("删除了","Revision:",event.Kv.ModRevision)
			}
		}
	}


}
