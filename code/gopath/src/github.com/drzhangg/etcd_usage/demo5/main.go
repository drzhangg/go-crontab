package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

func main() {
	var(
		config clientv3.Config
		client *clientv3.Client
		err error
		kv clientv3.KV
		delResp *clientv3.DeleteResponse
		kvpair *mvccpb.KeyValue
	)

	config = clientv3.Config{
		Endpoints:[]string{"47.99.240.52:2379"},
		DialTimeout:5 * time.Second,
	}
	
	//建立一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	//删除kv
	if delResp,err = kv.Delete(context.TODO(),"/cron/jobs/job2",clientv3.WithPrefix());err != nil {
		fmt.Println(err)
		return
	}

	//获取被删除之前的value
	if len(delResp.PrevKvs) != 0 {
		for _,kvpair = range delResp.PrevKvs{
			fmt.Println("删除了：",string(kvpair.Key),string(kvpair.Value))
		}
	}else {
		fmt.Println("nil:",len(delResp.PrevKvs))
	}


}
