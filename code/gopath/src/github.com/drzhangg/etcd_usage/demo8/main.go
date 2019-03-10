package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var(
		config clientv3.Config
		client *clientv3.Client
		err error
		kv clientv3.KV
		opPut clientv3.Op
		opResp clientv3.OpResponse
		opGet clientv3.Op
	)

	config = clientv3.Config{
		Endpoints:[]string{"47.99.240.52:2379"},
		DialTimeout:5 * time.Second,
	}

	//创建一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//kv
	kv = clientv3.NewKV(client)

	//创建Op：operation
	opPut = clientv3.OpPut("/cron/jobs/job9","test job9")

	//执行Op
	if opResp,err = kv.Do(context.TODO(),opPut);err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("读取Revision:",opResp.Put().Header.Revision)

	//创建Get op
	opGet = clientv3.OpGet("/cron/jobs/job9")

	//执行op
	if opResp,err = kv.Do(context.TODO(),opGet);err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("读取Revision:",opResp.Get().Kvs[0].ModRevision)
	fmt.Println("读取结果：",string(opResp.Get().Kvs[0].Value))
}